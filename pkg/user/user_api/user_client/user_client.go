package user_client

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type SetterHandler[T interface{}] struct {
	Cmd T
}

func (s *SetterHandler[T]) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {
	return client.Exec(ctx, operation, s.Cmd, nil)
}

type UserBuilder[U user.User] func() U

type UserClient[U user.User] struct {
	api_client.ServiceClient
	userTypeName string
	userBuilder  UserBuilder[U]

	collectionResource api.Resource

	add  api.Operation
	list api.Operation
}

func NewUserClient[U user.User](client api_client.Client,
	userBuilder UserBuilder[U],
	userTypeName ...string) *UserClient[U] {

	var serviceName string
	c := &UserClient[U]{}
	c.userTypeName, serviceName, c.collectionResource, _ = user_api.PrepareResources(userTypeName...)
	c.ServiceClient.Init(client, serviceName)

	c.Service().AddChild(c.collectionResource)
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

func (u *UserClient[U]) UserOperation(userId string, resourceName string, op api.Operation) api.Operation {
	userResource := user_api.NamedUserResource(userId, u.userTypeName)
	opResource := api.NewResource(resourceName)
	userResource.AddChild(opResource)
	opResource.AddOperation(op)
	u.Service().AddChild(userResource.Parent())
	return op
}

func (u *UserClient[U]) Find(ctx op_context.Context, id string) (U, error) {
	var nilU U
	return nilU, errors.New("not implemented yet")
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

func (u *UserClient[U]) SetBlocked(ctx op_context.Context, login string, blocked bool) error {
	return errors.New("not implemented yet")
}

func (u *UserClient[U]) FindUsers(ctx op_context.Context, filter *db.Filter, users *[]U) error {
	return errors.New("not implemented yet")
}
