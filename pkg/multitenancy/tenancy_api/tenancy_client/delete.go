package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Delete struct {
	tenancy_api.DeleteTenancyCmd
}

func (d *Delete) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("DeleteTenancy.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, d, nil)
	c.SetError(err)
	return err
}

func (t *TenancyClient) Delete(ctx op_context.Context, id string, withDb bool, idIsDisplay ...bool) error {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyClient.Delete")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// adjust ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get tenancy ID")
		return err
	}

	// prepare and exec handler
	handler := &Delete{}
	op := api.NamedResourceOperation(t.TenancyResource,
		tenancy_api.TenancyResource,
		tenancyId,
		tenancy_api.Delete())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
