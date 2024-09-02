package customer_api_client

import (
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/common"
	"github.com/evgeniums/go-utils/pkg/customer/customer_api"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/user/user_api/user_client"
)

type SetDescription = user_client.SetterHandler[common.WithDescriptionBase]

func (u *Client[T]) SetDescription(ctx op_context.Context, id string, description string, idIsLogin ...bool) error {

	// setup
	c := ctx.TraceInMethod("Client.SetDescription")
	defer ctx.TraceOutMethod()

	// if idIsLogin then first find user
	userId, err := u.GetUserId(ctx, id, idIsLogin...)
	if err != nil {
		c.SetMessage("failed to get user ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetDescription{}
	handler.Cmd.SetDescription(description)

	// prepare and exec handler
	err = u.UserOperation(userId, "description", customer_api.SetDescription()).Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil
}
