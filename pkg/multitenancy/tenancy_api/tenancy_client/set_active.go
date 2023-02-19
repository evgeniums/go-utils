package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type SetActive struct {
	cmd *common.WithActiveBase
}

func (s *SetActive) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("SetTenancyActive.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, s.cmd, nil)
	c.SetError(err)
	return err
}

func (t *TenancyClient) SetActive(ctx op_context.Context, id string, active bool, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetActive")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetActive{
		cmd: &common.WithActiveBase{ACTIVE: active},
	}

	// prepare and exec handler
	op := api.NamedResourceOperation(t.TenancyResource, "active", tenancyId, tenancy_api.SetActive())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil
}

func (t *TenancyClient) Activate(ctx op_context.Context, id string, idIsDisplay ...bool) error {
	return t.SetActive(ctx, id, true, idIsDisplay...)
}

func (t *TenancyClient) Deactivate(ctx op_context.Context, id string, idIsDisplay ...bool) error {
	return t.SetActive(ctx, id, false, idIsDisplay...)
}
