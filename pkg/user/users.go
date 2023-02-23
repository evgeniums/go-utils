package user

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const ErrorCodeDuplicateLogin string = "duplicate_login"
const ErrorCodeDuplicateEmail string = "duplicate_email"
const ErrorCodeDuplicatePhone string = "duplicate_phone"

var ErrorDescriptions = map[string]string{
	ErrorCodeDuplicateLogin: "Login already occupied.",
	ErrorCodeDuplicateEmail: "Email address already occupied.",
	ErrorCodeDuplicatePhone: "Phone number already occupied.",
}

var ErrorHttpCodes = map[string]int{}

type MainFieldSetters interface {
	SetPassword(ctx op_context.Context, id string, password string, idIsLogin ...bool) error
	SetPhone(ctx op_context.Context, id string, phone string, idIsLogin ...bool) error
	SetEmail(ctx op_context.Context, id string, email string, idIsLogin ...bool) error
	SetBlocked(ctx op_context.Context, id string, blocked bool, idIsLogin ...bool) error
}

type UserController[UserType User] interface {
	MainFieldSetters

	Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error)
	Find(ctx op_context.Context, id string) (UserType, error)
	FindByLogin(ctx op_context.Context, login string) (UserType, error)
	FindAuthUser(ctx op_context.Context, login string, user auth.User, dest ...interface{}) (bool, error)
	FindUsers(ctx op_context.Context, filter *db.Filter, users *[]UserType) (int64, error)

	SetUserBuilder(builder func() UserType)
	MakeUser() UserType

	SetOplogBuilder(builder func() OpLogUserI)
}

type Users[UserType User] interface {
	auth_session.AuthUserManager
	UserController[UserType]
}

type UserControllerBase[UserType User] struct {
	userBuilder    func() UserType
	oplogBuilder   func() OpLogUserI
	crudController crud.CRUD
	userValidators auth_session.UserValidators
}

func LocalUserController[UserType User]() *UserControllerBase[UserType] {
	return &UserControllerBase[UserType]{crudController: &crud.DbCRUD{}, oplogBuilder: func() OpLogUserI { return &OpLogUser{} }}
}

func (u *UserControllerBase[UserType]) SetUserBuilder(userBuilder func() UserType) {
	u.userBuilder = userBuilder
}

func (u *UserControllerBase[UserType]) SetOplogBuilder(builder func() OpLogUserI) {
	u.oplogBuilder = builder
}

func (u *UserControllerBase[UserType]) SetUserValidators(validators auth_session.UserValidators) {
	u.userValidators = validators
}

func (u *UserControllerBase[UserType]) CRUD() crud.CRUD {
	return u.crudController
}

func (u *UserControllerBase[UserType]) MakeUser() UserType {
	return u.userBuilder()
}

func (u *UserControllerBase[UserType]) Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error) {

	// setup
	var nilUser UserType
	ctx.SetLoggerField("login", login)
	c := ctx.TraceInMethod("Users.Add")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// validate
	if u.userValidators != nil {
		err = u.userValidators.ValidateLogin(login)
		if err != nil {
			c.SetMessage("failed to validate login")
			return nilUser, err
		}
		err = u.userValidators.ValidatePassword(password)
		if err != nil {
			c.SetMessage("failed to validate password")
			return nilUser, err
		}
	}

	// check if user with such login already exists
	var users []UserType
	filter := db.NewFilter()
	filter.AddField("login", login)
	filter.Limit = 1
	_, err = u.FindUsers(ctx, filter, &users)
	if err != nil {
		c.SetMessage("failed to check login duplicates")
		return nilUser, err
	}
	if len(users) > 0 {
		err = errors.New("duplicate login")
		ctx.SetGenericErrorCode(ErrorCodeDuplicateLogin)
		return nilUser, err
	}

	// create user
	user := u.MakeUser()
	user.InitObject()
	user.SetLogin(login)
	user.SetPassword(password)
	for _, setter := range extraFieldsSetters {
		checkDuplicateFields, err1 := setter(ctx, user)
		err = err1
		if err != nil {
			c.SetMessage("failed to set extra fields")
			return nilUser, err
		}
		for _, checkDup := range checkDuplicateFields {
			var users []UserType
			filter := db.NewFilter()
			filter.AddField(checkDup.Name, checkDup.Value)
			filter.Limit = 1
			_, err1 := u.FindUsers(ctx, filter, &users)
			err = err1
			if err != nil {
				c.SetLoggerField("unique_field", checkDup.Name)
				c.SetMessage("failed to check unique field")
				return nilUser, err
			}
			if len(users) > 0 {
				c.SetLoggerField("unique_field", checkDup.Name)
				c.SetLoggerField("unique_value", checkDup.Value)
				err = fmt.Errorf("duplicate %s", checkDup.Name)
				ctx.SetGenericErrorCode(checkDup.ErrorCode)
				return nilUser, err
			}
		}
	}

	// create in manager
	err = u.crudController.Create(ctx, user)
	if err != nil {
		c.SetMessage("failed to create user")
		return nilUser, err
	}
	u.OpLog(ctx, "add", user.GetID(), login)

	// done
	return user, nil
}

func (u *UserControllerBase[UserType]) Find(ctx op_context.Context, id string) (UserType, error) {

	var nilUser UserType

	// setup
	c := ctx.TraceInMethod("Users.Find", logger.Fields{"id": id})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user := u.MakeUser()
	found, err := u.crudController.Read(ctx, db.Fields{"id": id}, user)
	if err != nil {
		c.SetMessage("failed to find user in database")
		return nilUser, err
	}
	if !found {
		ctx.SetGenericErrorCode(generic_error.ErrorCodeNotFound)
		err = errors.New("user with such ID does not exist")
		return nilUser, err
	}

	// done
	return user, nil
}

func (u *UserControllerBase[UserType]) FindByLogin(ctx op_context.Context, login string) (UserType, error) {

	var nilUser UserType

	// setup
	c := ctx.TraceInMethod("Users.FindByLogin", logger.Fields{"login": login})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user := u.MakeUser()
	found, err := FindByLogin(u.crudController, ctx, login, user)
	if err != nil {
		c.SetMessage("failed to find user in database")
		return nilUser, err
	}
	if !found {
		err = errors.New("user with such login does not exist")
		return nilUser, err
	}

	// done
	return user, nil
}

func FindUser[UserType User](u *UserControllerBase[UserType], ctx op_context.Context, id string, idIsLogin ...bool) (user UserType, err error) {

	useLogin := utils.OptionalArg(false, idIsLogin...)

	if useLogin {
		ctx.SetLoggerField("login", id)
		user, err = u.FindByLogin(ctx, id)
	} else {
		ctx.SetLoggerField("id", id)
		user, err = u.Find(ctx, id)
	}

	return
}

func (u *UserControllerBase[UserType]) SetPassword(ctx op_context.Context, id string, password string, idIsLogin ...bool) error {

	// setup
	c := ctx.TraceInMethod("Users.SetPassword")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// validate
	if u.userValidators != nil {
		err = u.userValidators.ValidatePassword(password)
		if err != nil {
			c.SetMessage("failed to validate password")
			return err
		}
	}

	// find user
	user, err := FindUser(u, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set password
	user.SetPassword(password)
	err = u.crudController.Update(ctx, user, db.Fields{"password_hash": user.PasswordHash(), "password_salt": user.PasswordSalt()})
	if err != nil {
		return err
	}

	// done
	u.OpLog(ctx, "set_password", user.GetID(), user.Login())
	return nil
}

func (u *UserControllerBase[UserType]) SetPhone(ctx op_context.Context, id string, phone string, idIsLogin ...bool) error {

	// setup
	ctx.SetLoggerField("phone", phone)
	c := ctx.TraceInMethod("Users.SetPhone")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user, err := FindUser(u, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set phone
	err = u.crudController.Update(ctx, user, db.Fields{"phone": phone})
	if err != nil {
		return err
	}

	// done
	u.OpLog(ctx, "set_phone", user.GetID(), user.Login())
	return nil
}

func (u *UserControllerBase[UserType]) SetEmail(ctx op_context.Context, id string, email string, idIsLogin ...bool) error {

	// setup
	ctx.SetLoggerField("email", email)
	c := ctx.TraceInMethod("Users.SetEmail")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user, err := FindUser(u, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set email
	err = u.crudController.Update(ctx, user, db.Fields{"email": email})
	if err != nil {
		return err
	}

	// done
	u.OpLog(ctx, "set_email", user.GetID(), user.Login())
	return nil
}

func (u *UserControllerBase[UserType]) FindAuthUser(ctx op_context.Context, login string, user auth.User, dest ...interface{}) (bool, error) {
	return FindByLogin(u.crudController, ctx, login, user)
}

func (u *UserControllerBase[UserType]) SetBlocked(ctx op_context.Context, id string, blocked bool, idIsLogin ...bool) error {

	// setup
	ctx.SetLoggerField("blocked", blocked)
	c := ctx.TraceInMethod("Users.SetBlocked")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user, err := FindUser(u, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set blocked
	err = u.crudController.Update(ctx, user, db.Fields{"blocked": blocked})
	if err != nil {
		return err
	}

	// done
	u.OpLog(ctx, "set_blocked", user.GetID(), user.Login())
	return nil
}

func (u *UserControllerBase[UserType]) FindUsers(ctx op_context.Context, filter *db.Filter, users *[]UserType) (int64, error) {
	return u.crudController.List(ctx, filter, users)
}

func (u *UserControllerBase[UserType]) OpLog(ctx op_context.Context, op string, userId string, login string) {
	oplog := u.oplogBuilder()
	oplog.SetOperation(op)
	oplog.SetLogin(login)
	oplog.SetUserId(userId)
	ctx.Oplog(oplog)
}

type UsersValidator struct {
	Validator            validator.Validator
	LoginValidationRules string
}

func (u *UsersValidator) ValidateLogin(login string) error {
	err := u.Validator.ValidateValue(login, u.LoginValidationRules)
	if err != nil {
		return &validator.ValidationError{Message: "Invalid format of login", Field: "login"}
	}
	return nil
}

func (u *UsersValidator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return &validator.ValidationError{Message: "Password must be at least 8 characters", Field: "password"}
	}
	return nil
}

func (u *UsersValidator) Init(vld validator.Validator, loginValidationRules ...string) {
	u.Validator = vld
	u.LoginValidationRules = utils.OptionalArg("required,alphanum_|email,lowercase", loginValidationRules...)
}

type UsersBase[UserType User] struct {
	UsersValidator
	UserController[UserType]
}

func (u *UsersBase[UserType]) Construct(userController UserController[UserType]) {
	u.UserController = userController
}

func (u *UsersBase[UserType]) MakeAuthUser() auth.User {
	return u.MakeUser()
}

func (m *UsersBase[UserType]) AuthUserManager() auth_session.AuthUserManager {
	return m
}
