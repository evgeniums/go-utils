package user_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

func PrepareResources(userTypeName ...string) (userType string, serviceName string, collectionResource api.Resource, userResource api.Resource) {

	userType = utils.OptionalArg("user", userTypeName...)
	serviceName = utils.ConcatStrings(userType, "s")

	userResource = UserResource(userType)
	collectionResource = userResource.Parent()

	return
}

func NamedUserResource(id string, userTypeName ...string) (userResource api.Resource) {
	r := UserResource(userTypeName...)
	r.SetId(id)
	return r
}

func UserResource(resourceType ...string) api.Resource {
	return api.NamedResource(utils.OptionalArg("user", resourceType...))
}

type ListResponse[T any] struct {
	api.ResponseHateous
	Users *[]T `json:"users"`
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

func Find() api.Operation {
	return api.NewOperation("find", access_control.Get)
}

type SetPasswordCmd struct {
	Password string `json:"password"`
}

func SetPassword() api.Operation {
	return api.NewOperation("set_password", access_control.Put)
}

type SetEmailCmd struct {
	Email string `json:"email" validate:"omitempty,email" vmessage:"Invalid email format"`
}

func SetEmail() api.Operation {
	return api.NewOperation("set_email", access_control.Put)
}

type SetPhoneCmd struct {
	Phone string `json:"phone" validate:"omitempty,phone" vmessage:"Invalid phone format"`
}

func SetPhone() api.Operation {
	return api.NewOperation("set_phone", access_control.Put)
}

type SetBlockedCmd struct {
	Blocked bool `json:"blocked"`
}

func SetBlocked() api.Operation {
	return api.NewOperation("set_blocked", access_control.Put)
}
