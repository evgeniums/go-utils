package pool_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type ListServicePools struct {
	result *pool_api.ListServicePoolsResponse
}

func (a *ListServicePools) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("ListServicePools.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, a.result)
	c.SetError(err)
	return err
}

func (p *PoolClient) GetServiceBindings(ctx op_context.Context, id string, idIsName ...bool) ([]*pool.PoolServiceBinding, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.GetServiceBindings")
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
		return nil, err
	}

	// prepare and exec handler
	handler := &ListServicePools{
		result: &pool_api.ListServicePoolsResponse{},
	}
	resource := p.resourceForServicePools(sId)
	op := pool_api.ListServicePools()
	resource.AddOperation(op)
	err = op.Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.Items, nil
}
