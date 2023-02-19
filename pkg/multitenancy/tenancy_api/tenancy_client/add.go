package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Add struct {
	cmd    *multitenancy.TenancyData
	result *tenancy_api.TenancyResponse
}

func (a *Add) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("AddTenancy.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, a.cmd, a.result)
	c.SetError(err)
	return err
}

func (t *TenancyClient) Add(ctx op_context.Context, tenancy *multitenancy.TenancyData) (*multitenancy.TenancyItem, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyClient.Add")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare and exec handler
	handler := &Add{
		cmd:    tenancy,
		result: &tenancy_api.TenancyResponse{},
	}
	err = t.add.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.TenancyItem, nil
}
