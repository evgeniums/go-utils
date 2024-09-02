package tenancy_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/common"
	"github.com/evgeniums/go-utils/pkg/multitenancy"
	"github.com/evgeniums/go-utils/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-utils/pkg/op_context"
)

func (t *TenancyClient) SetActive(ctx op_context.Context, id string, active bool, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetActive")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := multitenancy.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := api_client.NewHandlerCmd(&common.WithActiveBase{ACTIVE: active})

	// prepare and exec handler
	op := api.OperationAsResource(t.TenancyResource, "active", tenancyId, tenancy_api.SetActive())
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
