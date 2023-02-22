package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type SetCustomer struct {
	cmd *multitenancy.WithCustomerId
}

func (s *SetCustomer) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("SetTenancyCustomer.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, s.cmd, nil)
	c.SetError(err)
	return err
}

func (t *TenancyClient) SetCustomer(ctx op_context.Context, id string, customerId string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetCustomer")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetCustomer{
		cmd: &multitenancy.WithCustomerId{CUSTOMER_ID: customerId},
	}

	// prepare and exec handler
	op := api.OperationAsResource(t.TenancyResource, "customer", tenancyId, tenancy_api.SetCustomer())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil

}
