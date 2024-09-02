package pool_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type AddService struct {
	cmd    pool.PoolService
	result *pool_api.ServiceResponse
}

func (a *AddService) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("AddService.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, a.cmd, a.result)
	c.SetError(err)
	return err
}

func (p *PoolClient) AddService(ctx op_context.Context, service pool.PoolService) (pool.PoolService, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.AddService")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare and exec handler
	handler := &AddService{
		cmd:    service,
		result: &pool_api.ServiceResponse{},
	}
	err = p.add_service.Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.PoolServiceBase, nil
}
