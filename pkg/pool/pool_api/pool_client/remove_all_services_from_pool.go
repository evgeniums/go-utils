package pool_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type RemoveAllServicesFromPool struct{}

func (a *RemoveAllServicesFromPool) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("ListPoolservices.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, nil)
	c.SetError(err)
	return err
}

func (p *PoolClient) RemoveAllServicesFromPool(ctx op_context.Context, id string, idIsName ...bool) error {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.RemoveAllServicesFromPool")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// adjust pool ID
	pId, _, err := p.poolId(ctx, id, idIsName...)
	if err != nil {
		return err
	}

	// prepare and exec handler
	handler := &RemoveAllServicesFromPool{}
	resource := p.resourceForPoolServices(pId)
	op := pool_api.RemoveAllServicesFromPool()
	resource.AddOperation(op)
	err = op.Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// done
	return nil
}
