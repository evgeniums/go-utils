package user_client

import (
	"errors"

	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/auth"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/user"
	"github.com/evgeniums/go-utils/pkg/user/user_api"
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

	CollectionResource api.Resource
	UserResource       api.Resource

	add  api.Operation
	list api.Operation
}

func NewUserClient[U user.User](client api_client.Client,
	userBuilder UserBuilder[U],
	userTypeName ...string) *UserClient[U] {

	var serviceName string
	c := &UserClient[U]{}
	c.userTypeName, serviceName, c.CollectionResource, c.UserResource = user_api.PrepareResources(userTypeName...)
	c.ServiceClient.Init(client, serviceName)

	c.AddChild(c.CollectionResource)
	c.add = user_api.Add(c.userTypeName)
	c.list = user_api.List()
	c.CollectionResource.AddOperations(c.add, c.list)

	return c
}

func (c *UserClient[U]) OpLog(ctx op_context.Context, op string, userId string, login string) {}

func (c *UserClient[U]) SetTenancy(tenancyResource api.Resource) {
	tenancyResource.AddChild(c)
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
	userResource := u.UserResource.CloneChain(false)
	userResource.SetId(userId)
	userResource.AddChild(opResource)
	return op
}

func (u *UserClient[U]) FindAuthUser(ctx op_context.Context, login string) (auth.User, error) {
	return nil, errors.New("unsupported method")
}
