package pool_service

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type UpdateServiceEndpoint struct {
	PoolEndpoint
}

func (e *UpdateServiceEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.UpdateService")
	defer request.TraceOutMethod()

	// parse command
	cmd := &api.UpdateCmd{}
	err := request.ParseVerify(cmd)
	if err != nil {
		c.SetMessage("faield to parse/validate command")
		return err
	}
	// validate fields
	err = validator.ValidateMap(request.App().Validator(), cmd.Fields, &pool.PoolServiceBaseEssentials{})
	if err != nil {
		c.SetMessage("faield to validate fields")
		return err
	}

	// update service
	serviceId := request.GetResourceId("service")
	err = e.service.Pools.UpdateService(request, serviceId, cmd.Fields)
	if err != nil {
		c.SetMessage("failed to update service")
		return c.SetError(err)
	}

	// find updated service
	s, err := e.service.Pools.FindService(request, serviceId)
	if err != nil {
		c.SetMessage("failed to find service")
		return c.SetError(err)
	}
	if s == nil {
		request.SetGenericErrorCode(pool.ErrorCodeServiceNotFound)
		return c.SetError(errors.New("service not found"))
	}

	// set response
	resp := &pool_api.ServiceResponse{}
	resp.PoolServiceBase = s.(*pool.PoolServiceBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func UpdateService(s *PoolService) *UpdateServiceEndpoint {
	e := &UpdateServiceEndpoint{}
	e.Construct(s, api.UpdatePartial())
	return e
}
