package pool_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type RemoveAllServicesFromPoolEndpoint struct {
	PoolEndpoint
}

func (e *RemoveAllServicesFromPoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.RemoveAllServicesFromPool")
	defer request.TraceOutMethod()

	// do operation
	poolId := request.GetResourceId("pool")
	err := e.service.Pools.RemoveAllServicesFromPool(request, poolId)
	if err != nil {
		c.SetMessage("failed to remove services from pool")
		return c.SetError(err)
	}

	// done
	return nil
}

func RemoveAllServicesFromPool(s *PoolService) *RemoveAllServicesFromPoolEndpoint {
	e := &RemoveAllServicesFromPoolEndpoint{}
	e.Construct(s, pool_api.RemoveAllServicesFromPool())
	return e
}
