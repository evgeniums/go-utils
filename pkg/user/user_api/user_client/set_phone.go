package user_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type SetPhone = SetterHandler[user_api.SetPhoneCmd]

func (u *UserClient[U]) SetPhone(ctx op_context.Context, id string, phone string, idIsLogin ...bool) error {

	// setup
	c := ctx.TraceInMethod("UserClient.SetPhone")
	defer ctx.TraceOutMethod()

	// TODO if idIsLogin then first find user

	// create command
	handler := &SetPhone{}
	handler.Cmd.Phone = phone

	// prepare and exec handler
	err := u.UserOperation(id, "phone", user_api.SetPhone()).Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
