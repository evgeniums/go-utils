package pool_service

import (
	"errors"

	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
	"github.com/evgeniums/go-utils/pkg/validator"
)

type UpdatePoolEndpoint struct {
	PoolEndpoint
}

func (e *UpdatePoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.UpdatePool")
	defer request.TraceOutMethod()

	// parse command
	cmd := &api.UpdateCmd{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return c.SetError(err)
	}
	// validate fields
	vErr := validator.ValidateMap(request.App().Validator(), cmd.Fields, &pool.PoolBaseData{})
	if vErr != nil {
		c.SetMessage("failed to validate fields")
		request.SetGenericError(vErr.GenericError())
		return c.SetError(vErr.Err)
	}

	// update pool
	poolId := request.GetResourceId("pool")
	p, err := e.service.Pools.UpdatePool(request, poolId, cmd.Fields)
	if err != nil {
		c.SetMessage("failed to update pool")
		return c.SetError(err)
	}

	// find updated pool
	if p == nil {
		request.SetGenericErrorCode(pool.ErrorCodePoolNotFound)
		return c.SetError(errors.New("pool not found"))
	}

	// set response
	resp := &pool_api.PoolResponse{}
	resp.PoolBase = p.(*pool.PoolBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func UpdatePool(s *PoolService) *UpdatePoolEndpoint {
	e := &UpdatePoolEndpoint{}
	e.Construct(s, pool_api.UpdatePool())
	return e
}
