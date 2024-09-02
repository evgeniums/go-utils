package pool_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type FindService struct {
	result *pool_api.ServiceResponse
}

func (a *FindService) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("FindService.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, a.result)
	c.SetError(err)
	return err
}

func (p *PoolClient) FindService(ctx op_context.Context, id string, idIsName ...bool) (pool.PoolService, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.FindService")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	sId, service, err := p.serviceId(ctx, id, idIsName...)
	if err != nil {
		return nil, err
	}
	if service != nil {
		return service, nil
	}

	// prepare and exec handler
	handler := &FindService{
		result: &pool_api.ServiceResponse{},
	}
	err = api.NamedResourceOperation(p.ServiceResource, sId, pool_api.FindService()).Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.PoolServiceBase, nil
}
