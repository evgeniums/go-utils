package user_client

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type UserBuilder[U user.User] func() U

type UserClient[U user.User] struct {
	api_client.ServiceClient
	userTypeName string
	userBuilder  UserBuilder[U]

	collectionResource api.Resource
	userResource       api.Resource

	add  api.Operation
	list api.Operation
}

func NewUserClient[U user.User](client api_client.Client,
	userBuilder UserBuilder[U],
	userTypeName ...string) *UserClient[U] {

	var serviceName string
	c := &UserClient[U]{}
	c.userTypeName, serviceName, c.collectionResource, c.userResource = user_api.PrepareResources(userTypeName...)
	c.ServiceClient.Init(client, serviceName)

	c.AddChild(c.collectionResource)
	c.add = user_api.Add()
	c.list = user_api.List()

	c.collectionResource.AddOperations(c.add, c.list)

	return c
}

func (c *UserClient[U]) SetUserBuilder(userBuilder func() U) {
	c.userBuilder = userBuilder
}

func (c *UserClient[U]) MakeUser() U {
	return c.userBuilder()
}

type Add[U user.User] struct {
	cmd    interface{}
	result *user_api.UserResponse[U]
}

func (a *Add[U]) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("Add.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, a.cmd, a.result)
	c.SetError(err)
	return err
}

func (u *UserClient[U]) Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...user.SetUserFields[U]) (U, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("UserClient.Add", logger.Fields{"login": login, "user_type": u.userTypeName})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	var nilU U

	// create user
	user := u.userBuilder()
	user.SetLogin(login)
	for _, setter := range extraFieldsSetters {
		err := setter(ctx, user)
		if err != nil {
			c.SetMessage("failed to set extra field")
			return nilU, err
		}
	}

	// create command from user
	cmd := user.ToCmd(password)

	// prepare and exec handler
	handler := &Add[U]{
		cmd:    cmd,
		result: &user_api.UserResponse[U]{},
	}
	err = u.add.Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nilU, err
	}

	// return result
	return handler.result.User, nil
}

func (u *UserClient[U]) FindByLogin(ctx op_context.Context, login string) (U, error) {
	var nilU U
	return nilU, errors.New("not implemented yet")
}

func (u *UserClient[U]) FindAuthUser(ctx op_context.Context, login string, user auth.User, dest ...interface{}) (bool, error) {
	return false, errors.New("unsupported method")
}

func (u *UserClient[U]) SetPassword(ctx op_context.Context, login string, password string) error {
	return errors.New("not implemented yet")
}

func (u *UserClient[U]) SetEmail(ctx op_context.Context, login string, email string) error {
	return errors.New("not implemented yet")
}

func (u *UserClient[U]) SetPhone(ctx op_context.Context, login string, phone string) error {
	return errors.New("not implemented yet")
}

func (u *UserClient[U]) SetBlocked(ctx op_context.Context, login string, blocked bool) error {
	return errors.New("not implemented yet")
}

func (u *UserClient[U]) FindUsers(ctx op_context.Context, filter *db.Filter, users *[]U) error {
	return errors.New("not implemented yet")
}
