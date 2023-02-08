package user_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

/*

type UserController[UserType User] interface {

	Find(ctx op_context.Context, fields db.Fields, user interface{}) (bool, error)
	Create(ctx op_context.Context, user common.Object) error
	List(ctx op_context.Context, filter *db.Filter, users interface{}) error
	Add(ctx op_context.Context, login string, password string, extraFieldsSetters ...SetUserFields[UserType]) (UserType, error)
	FindByLogin(ctx op_context.Context, login string) (UserType, error)

	Update(ctx op_context.Context, user common.Object, fields db.Fields) error
	SetPassword(ctx op_context.Context, login string, password string) error
	SetPhone(ctx op_context.Context, login string, phone string) error
	SetEmail(ctx op_context.Context, login string, email string) error
	SetBlocked(ctx op_context.Context, login string, blocked bool) error

	SetUserBuilder(builder func() UserType)
	MakeUser() UserType
}

*/

func PrepareResources(userTypeName ...string) (serviceName string, groupResource api.Resource, userResource api.Resource) {

	userType := utils.OptionalArg("user", userTypeName...)
	serviceName = utils.ConcatStrings(userType, "s")

	userResource = UserResource(userType)
	groupResource = userResource.Parent()

	return
}

func UserResource(resourceType ...string) api.Resource {
	return api.NamedResource(utils.OptionalArg("user", resourceType...))
}

func Find() api.Operation {
	return api.NewOperation("find", access_control.Read)
}

func Create() api.Operation {
	return api.NewOperation("create", access_control.Create)
}

func Update() api.Operation {
	return api.NewOperation("update", access_control.UpdatePartial)
}
