package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type SetRole struct {
	cmd *multitenancy.WithRole
}

func (s *SetRole) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("SetTenancyRole.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, s.cmd, nil)
	c.SetError(err)
	return err
}

func (t *TenancyClient) SetRole(ctx op_context.Context, id string, role string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetRole")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetRole{
		cmd: &multitenancy.WithRole{ROLE: role},
	}

	// prepare and exec handler
	op := api.NamedResourceOperation(t.TenancyResource, "role", tenancyId, tenancy_api.SetRole())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil

}
