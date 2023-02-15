package pool_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
)

type RemoveServiceFromAllPools struct{}

func (a *RemoveServiceFromAllPools) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("RemoveServiceFromAllPools.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, nil)
	c.SetError(err)
	return err
}

func (p *PoolClient) RemoveServiceFromAllPools(ctx op_context.Context, id string, idIsName ...bool) error {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.RemoveServiceFromAllPools")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// adjust service ID
	sId, _, err := p.serviceId(ctx, id, idIsName...)
	if err != nil {
		return err
	}

	// prepare and exec handler
	handler := &RemoveServiceFromAllPools{}
	resource := p.resourceForServicePools(sId)
	op := pool_api.RemoveServiceFromAllPools()
	resource.AddOperation(op)
	err = op.Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
