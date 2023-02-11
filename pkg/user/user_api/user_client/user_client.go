package user_client

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
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

func (c *UserClient[U]) SetOplogBuilder(userBuilder func() user.OpLogUserI) {
}

func (c *UserClient[U]) MakeUser() U {
	return c.userBuilder()
}

func (u *UserClient[U]) UserOperation(userId string, resourceName string, op api.Operation) api.Operation {
	opResource := api.NewResource(resourceName)
	opResource.AddOperation(op)
	userResource := u.userResource.CloneChain(false)
	userResource.SetId(userId)
	userResource.AddChild(opResource)
	return op
}

func (u *UserClient[U]) FindAuthUser(ctx op_context.Context, login string, user auth.User, dest ...interface{}) (bool, error) {
	return false, errors.New("unsupported method")
}
