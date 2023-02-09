package user_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type UserCmdBuilder[U user.User] func(ctx op_context.Context, login string, password string, extraFieldsSetters ...user.SetUserFields[user.UserFieldsSetter[U]]) (user.UserFieldsSetter[U], error)

type UserClient[U user.User] struct {
	api_client.ServiceClient
	userTypeName string
	userBuilder  func() U

	collectionResource api.Resource
	userResource       api.Resource

	userCmdBuilder UserCmdBuilder[U]

	add  api.Operation
	list api.Operation
}

func NewUserClient[U user.User](client api_client.Client,
	userCmdBuilder UserCmdBuilder[U],
	userTypeName ...string) *UserClient[U] {

	var serviceName string
	c := &UserClient[U]{}
	c.userTypeName, serviceName, c.collectionResource, c.userResource = user_api.PrepareResources(userTypeName...)
	c.Init(client, serviceName)
	c.userCmdBuilder = userCmdBuilder

	c.AddChild(c.collectionResource)
	c.add = user_api.Add()
	c.list = user_api.List()

	c.collectionResource.AddOperations(c.add, c.list)

	// users.AddOperation(List(s))
	// users.AddOperation(Add(s, setterBuilder))

	// user.AddOperation(Update(s))

	return c
}

func (c *UserClient[U]) SetUserBuilder(userBuilder func() U) {
	c.userBuilder = userBuilder
}

func (c *UserClient[U]) MakeUser() U {
	return c.userBuilder()
}

type Add[U user.User] struct {
	cmd    user.UserFieldsSetter[U]
	result *user_api.UserResponse[U]
}

func (a *Add[U]) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) generic_error.Error {
	return client.Exec(ctx, operation, a.cmd, a.result)
}

func (c *UserClient[U]) Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...user.SetUserFields[user.UserFieldsSetter[U]]) (U, error) {

	var nilU U

	cmd, err := c.userCmdBuilder(ctx, login, password, extraFieldsSetters...)
	if err != nil {
		return nilU, err
	}
	handler := &Add[U]{
		cmd:    cmd,
		result: &user_api.UserResponse[U]{},
	}
	err = c.add.Exec(ctx, api_client.MakeOperationHandler(c.Client, handler))
	if err != nil {
		return nilU, err
	}

	return handler.result.User, nil
}

/*
	FindByLogin(ctx op_context.Context, login string) (UserType, error)
	FindAuthUser(ctx op_context.Context, login string, user interface{}, dest ...interface{}) (bool, error)
	SetPassword(ctx op_context.Context, login string, password string) error
	SetPhone(ctx op_context.Context, login string, phone string) error
	SetEmail(ctx op_context.Context, login string, email string) error
	SetBlocked(ctx op_context.Context, login string, blocked bool) error

	// TODO paginate users
	FindUsers(ctx op_context.Context, filter *db.Filter, users *[]UserType) error

*/
