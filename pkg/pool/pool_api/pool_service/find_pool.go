package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
)

type FindPoolEndpoint struct {
	PoolEndpoint
}

func (e *FindPoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.FindPool")
	defer request.TraceOutMethod()

	// find pool
	p, err := e.service.Pools.FindPool(request, request.GetResourceId("pool"))
	if err != nil {
		c.SetMessage("failed to find pool")
		return c.SetError(err)
	}

	// find pool services
	services, err := e.service.Pools.GetPoolBindings(request, p.GetID())
	if err != nil {
		c.SetMessage("failed to get pool services")
		return c.SetError(err)
	}
	p.SetServices(services)

	// set response
	resp := &pool_api.PoolResponse{}
	resp.PoolBase = p.(*pool.PoolBase)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func FindPool(s *PoolService) *FindPoolEndpoint {
	e := &FindPoolEndpoint{}
	e.Construct(s, pool_api.FindPool())
	return e
}
