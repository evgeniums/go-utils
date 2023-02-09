package admin_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/admin/admin_api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_client"
)

func MakeAdminCmd(ctx op_context.Context, login string, password string, extraFieldsSetters ...user.SetUserFields[user.UserFieldsSetter[*admin.Admin]]) (user.UserFieldsSetter[*admin.Admin], error) {

	c := ctx.TraceInMethod("MakeAdminFieldSetter")
	defer ctx.TraceOutMethod()

	a := &admin_api.AdminFieldsSetter{}

	a.SetLogin(login)
	a.SetPassword(password)
	for _, setter := range extraFieldsSetters {
		err := setter(ctx, a)
		if err != nil {
			c.SetMessage("failed to set extra fields")
		}
	}

	return a, nil
}

type AdminClient = user_client.UserClient[*admin.Admin]

func NewAdminClient(client api_client.Client) *AdminClient {
	return user_client.NewUserClient(client, MakeAdminCmd, "admin")
}
