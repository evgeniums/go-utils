package user_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type SetPhone user_api.SetPhoneCmd

func (s *SetPhone) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("SetPhone.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, s, nil)
	c.SetError(err)
	return err
}

func (u *UserClient[U]) SetPhone(ctx op_context.Context, id string, phone string) error {

	// setup
	c := ctx.TraceInMethod("UserClient.SetPhone")
	defer ctx.TraceOutMethod()

	// create command
	handler := &SetPhone{}
	handler.Phone = phone

	// prepare and exec handler
	err := u.UserOperation(id, "phone", user_api.SetPhone()).Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
