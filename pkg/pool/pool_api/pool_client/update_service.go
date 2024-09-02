package pool_client

import (
	"errors"

	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
	"github.com/evgeniums/go-utils/pkg/utils"
)

type UpdateService struct {
	cmd    *api.UpdateCmd
	result *pool_api.ServiceResponse
}

func (a *UpdateService) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("UpdateService.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, a.cmd, a.result)
	c.SetError(err)
	return err
}

func (p *PoolClient) UpdateService(ctx op_context.Context, id string, fields db.Fields, idIsName ...bool) (pool.PoolService, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("PoolClient.UpdateService")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// adjust id
	sId, service, err := p.serviceId(ctx, id, idIsName...)
	if err != nil {
		return nil, err
	}
	if utils.OptionalArg(false, idIsName...) && service == nil {
		// service not found by name
		ctx.SetGenericErrorCode(pool.ErrorCodeServiceNotFound)
		return nil, errors.New("service not found by name")
	}

	// prepare and exec handler
	handler := &UpdateService{
		cmd:    &api.UpdateCmd{},
		result: &pool_api.ServiceResponse{},
	}
	handler.cmd.Fields = fields
	err = api.NamedResourceOperation(p.ServiceResource, sId, pool_api.UpdateService()).Exec(ctx, api_client.MakeOperationHandler(p.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, err
	}

	// done
	return handler.result.PoolServiceBase, nil
}
