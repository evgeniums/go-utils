package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
)

type AddServiceEndpoint struct {
	PoolEndpoint
}

func (e *AddServiceEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.AddService")
	defer request.TraceOutMethod()

	// parse command
	cmd := pool.NewPoolService()
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("faield to parse/validate command")
		return err
	}

	// add service
	s, err := e.service.Pools.AddService(request, cmd)
	if err != nil {
		c.SetMessage("failed to add service")
		return c.SetError(err)
	}

	// set response
	resp := &pool_api.ServiceResponse{}
	resp.PoolServiceBase = s.(*pool.PoolServiceBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func AddService(s *PoolService) *AddServiceEndpoint {
	e := &AddServiceEndpoint{}
	e.Construct(s, api.Add())
	return e
}
