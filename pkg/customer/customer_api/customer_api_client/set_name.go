package customer_api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/customer/customer_api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_client"
)

type SetName = user_client.SetterHandler[common.WithNameBase]

func (u *CustomerClient) SetName(ctx op_context.Context, id string, name string, idIsLogin ...bool) error {

	// setup
	c := ctx.TraceInMethod("CustomerClient.SetName")
	defer ctx.TraceOutMethod()

	// if idIsLogin then first find user
	userId, err := u.GetUserId(ctx, id, idIsLogin...)
	if err != nil {
		c.SetMessage("failed to get user ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetName{}
	handler.Cmd.SetName(name)

	// prepare and exec handler
	err = u.UserOperation(userId, "name", customer_api.SetName()).Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil
}
