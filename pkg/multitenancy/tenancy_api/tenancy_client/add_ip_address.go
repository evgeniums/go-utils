package tenancy_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/multitenancy"
	"github.com/evgeniums/go-utils/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-utils/pkg/op_context"
)

func (t *TenancyClient) AddIpAddress(ctx op_context.Context, id string, ipAddress string, tag string, idIsDisplay ...bool) error {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyClient.AddIpAddress")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// adjust ID
	tenancyId, _, err := multitenancy.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get tenancy ID")
		return err
	}

	// prepare and exec handler
	cmd := &tenancy_api.IpAddressCmd{
		Ip:  ipAddress,
		Tag: tag,
	}
	handler := api_client.NewHandlerCmd(cmd)
	op := api.OperationAsResource(t.TenancyResource, tenancy_api.IpAddressResource, tenancyId, tenancy_api.AddIpAddress())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
