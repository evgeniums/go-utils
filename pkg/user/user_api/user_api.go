package user_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

/*

type UserController[UserType User] interface {

	List(ctx op_context.Context, filter *db.Filter, users interface{}) error
	Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error)
	FindByLogin(ctx op_context.Context, login string) (UserType, error)

	SetPassword(ctx op_context.Context, login string, password string) error
	SetPhone(ctx op_context.Context, login string, phone string) error
	SetEmail(ctx op_context.Context, login string, email string) error
	SetBlocked(ctx op_context.Context, login string, blocked bool) error

	SetUserBuilder(builder func() UserType)
	MakeUser() UserType
}

*/

func PrepareResources(userTypeName ...string) (userType string, serviceName string, collectionResource api.Resource, userResource api.Resource) {

	userType = utils.OptionalArg("user", userTypeName...)
	serviceName = utils.ConcatStrings(userType, "s")

	userResource = UserResource(userType)
	collectionResource = userResource.Parent()

	return
}

func UserResource(resourceType ...string) api.Resource {
	return api.NamedResource(utils.OptionalArg("user", resourceType...))
}

type ListResponse[T any] struct {
	api.ResponseHateous
	Users []T `json:"users"`
}

func List() api.Operation {
	return api.NewOperation("list", access_control.Read)
}

type UserResponse[T user.User] struct {
	api.ResponseHateous
	User T `json:"user"`
}

func Add() api.Operation {
	return api.NewOperation("add", access_control.Create)
}
