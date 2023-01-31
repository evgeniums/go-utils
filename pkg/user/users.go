package user

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type Users[UserType User] struct {
	app_context.WithAppBase
	user_manager.UserManagerBase
	MakeUser func() UserType
}

func NewUsersTemplate[UserType User](makeUser func() UserType) *Users[UserType] {
	u := &Users[UserType]{}
	u.MakeUser = makeUser
	return u
}

func (u *Users[UserType]) Init(app app_context.Context) {
	u.WithAppBase.Init(app)
}

func (u *Users[UserType]) MakeAuthUser() auth.User {
	return u.MakeUser()
}

func (u *Users[UserType]) ValidateLogin(login string) error {
	rules := "required,alphanum,lowercase"
	return u.App().Validator().ValidateValue(login, rules)
}

func (u *Users[UserType]) Add(ctx op_context.Context, login string, password string, setExtraFields ...func(ctx op_context.Context, user UserType) error) error {

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
	if len(setExtraFields) > 0 {
		err = setExtraFields[0](ctx, user)
		if err != nil {
			c.SetMessage("failed to set extra fields")
		}
	}

	// create in manager
	err = u.Create(ctx, user)
	if err != nil {
		return err
	}

	// done
	return nil
}

func (u *Users[UserType]) FindByLogin(ctx op_context.Context, login string) (UserType, error) {

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
	var user UserType
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

func (u *Users[UserType]) SetPassword(ctx op_context.Context, login string, password string) error {

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

func (u *Users[UserType]) SetPhone(ctx op_context.Context, login string, phone string) error {

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

func (u *Users[UserType]) SetBlocked(ctx op_context.Context, login string, blocked bool) error {

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

func (m *Users[UserType]) UserManager() user_manager.UserManager {
	return m
}

type UsersWithSession[UserType User, SessionType user_manager.Session, SessionClientType user_manager.SessionClient] struct {
	Users[UserType]
	user_manager.SessionManagerBase
}

func NewUsersWithSessionTemplate[UserType User, SessionType user_manager.Session, SessionClientType user_manager.SessionClient](
	makeUser func() UserType, makeSession func() user_manager.Session, makeSessionClient func() user_manager.SessionClient) *UsersWithSession[UserType, SessionType, SessionClientType] {
	u := &UsersWithSession[UserType, SessionType, SessionClientType]{}
	u.MakeUser = makeUser
	u.MakeSession = makeSession
	u.MakeSessionClient = makeSessionClient
	return u
}

func (m *UsersWithSession[UserType, SessionType, SessionClientType]) SessionManager() user_manager.SessionManager {
	return m
}

type UsersBase = UsersWithSession[*UserBase, *user_manager.SessionBase, *user_manager.SessionClientBase]

func NewUsers() *UsersBase {
	return NewUsersWithSessionTemplate[*UserBase, *user_manager.SessionBase, *user_manager.SessionClientBase](NewUser, user_manager.NewSession, user_manager.NewSessionClient)
}
