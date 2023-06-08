package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func (t *TenancyClient) SetDbRole(ctx op_context.Context, id string, role string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetDbRole")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := multitenancy.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := api_client.NewHandlerCmd(&multitenancy.WithRole{ROLE: role})

	// prepare and exec handler
	op := api.OperationAsResource(t.TenancyResource, "db_role", tenancyId, tenancy_api.SetDbRole())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil

}
