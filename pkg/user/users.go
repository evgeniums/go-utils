package user

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type UserController[UserType User] interface {
	user_manager.UserController

	Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error)
	FindByLogin(ctx op_context.Context, login string) (UserType, error)
	SetPassword(ctx op_context.Context, login string, password string) error
	SetPhone(ctx op_context.Context, login string, phone string) error
	SetEmail(ctx op_context.Context, login string, email string) error
	SetBlocked(ctx op_context.Context, login string, blocked bool) error

	FindUsers(ctx op_context.Context, filter *db.Filter, users *[]UserType) error

	SetUserBuilder(builder func() UserType)
	MakeUser() UserType
}

type UserControllerBase[UserType User] struct {
	user_manager.UserControllerBase
	userBuilder func() UserType
}

func LocalUserController[UserType User]() *UserControllerBase[UserType] {
	return &UserControllerBase[UserType]{}
}

func (u *UserControllerBase[UserType]) SetUserBuilder(userBuilder func() UserType) {
	u.userBuilder = userBuilder
}

func (u *UserControllerBase[UserType]) MakeUser() UserType {
	return u.userBuilder()
}

func (u *UserControllerBase[UserType]) Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error) {

	// setup
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

	// create user
	user := u.MakeUser()
	user.InitObject()
	user.SetLogin(login)
	user.SetPassword(password)
	for _, setter := range extraFieldsSetters {
		err = setter(ctx, user)
		if err != nil {
			c.SetMessage("failed to set extra fields")
		}
	}

	// create in manager
	err = u.Create(ctx, user)
	if err != nil {
		var nilUser UserType
		return nilUser, err
	}

	// done
	return user, nil
}

func (u *UserControllerBase[UserType]) FindByLogin(ctx op_context.Context, login string) (UserType, error) {

	var nilUser UserType

	// setup
	c := ctx.TraceInMethod("Users.Find", logger.Fields{"login": login})
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
	found, err := user_manager.FindByLogin(u, ctx, login, user)
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

func (u *UserControllerBase[UserType]) SetPassword(ctx op_context.Context, login string, password string) error {

	// setup
	ctx.SetLoggerField("login", login)
	c := ctx.TraceInMethod("Users.SetPassword")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find admin
	user, err := u.FindByLogin(ctx, login)
	if err != nil {
		return err
	}

	// set password
	user.SetPassword(password)
	err = u.Update(ctx, user, db.Fields{"password_hash": user.PasswordHash(), "password_salt": user.PasswordSalt()})
	if err != nil {
		return err
	}

	// done
	return nil
}

func (u *UserControllerBase[UserType]) SetPhone(ctx op_context.Context, login string, phone string) error {

	// setup
	ctx.SetLoggerField("login", login)
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
	user, err := u.FindByLogin(ctx, login)
	if err != nil {
		return err
	}

	// set password
	err = u.Update(ctx, user, db.Fields{"phone": phone})
	if err != nil {
		return err
	}

	// done
	return nil
}

func (u *UserControllerBase[UserType]) SetEmail(ctx op_context.Context, login string, email string) error {

	// setup
	ctx.SetLoggerField("login", login)
	ctx.SetLoggerField("phone", email)
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
	user, err := u.FindByLogin(ctx, login)
	if err != nil {
		return err
	}

	// set password
	err = u.Update(ctx, user, db.Fields{"email": email})
	if err != nil {
		return err
	}

	// done
	return nil
}

func (u *UserControllerBase[UserType]) SetBlocked(ctx op_context.Context, login string, blocked bool) error {

	// setup
	ctx.SetLoggerField("login", login)
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

	// find admin
	user, err := u.FindByLogin(ctx, login)
	if err != nil {
		return err
	}

	// set password
	err = u.Update(ctx, user, db.Fields{"blocked": blocked})
	if err != nil {
		return err
	}

	// done
	return nil
}

func (u *UserControllerBase[UserType]) FindUsers(ctx op_context.Context, filter *db.Filter, users *[]UserType) error {
	return u.List(ctx, filter, users)
}

type UsersBase[UserType User] struct {
	app_context.WithAppBase
	LoginValidationRules string

	UserController[UserType]
}

func (u *UsersBase[UserType]) Construct(userController UserController[UserType]) {
	u.UserController = userController
}

func (u *UsersBase[UserType]) Init(app app_context.Context, loginValidationRules ...string) {
	u.WithAppBase.Init(app)
	u.LoginValidationRules = utils.OptionalArg("required,alphanum_|email,lowercase", loginValidationRules...)
}

func (u *UsersBase[UserType]) MakeAuthUser() auth.User {
	return u.MakeUser()
}

func (u *UsersBase[UserType]) ValidateLogin(login string) error {
	return u.App().Validator().ValidateValue(login, u.LoginValidationRules)
}

func (m *UsersBase[UserType]) UserManager() user_manager.UserManager {
	return m
}
