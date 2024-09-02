package pool_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type RemoveServiceFromPoolEndpoint struct {
	PoolEndpoint
}

func (e *RemoveServiceFromPoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.RemoveServiceToPool")
	defer request.TraceOutMethod()

	// do operation
	poolId := request.GetResourceId("pool")
	role := request.GetResourceId("role")
	err := e.service.Pools.RemoveServiceFromPool(request, poolId, role)
	if err != nil {
		c.SetMessage("failed to remove service from pool")
		return c.SetError(err)
	}

	// done
	return nil
}

func RemoveServiceFromPool(s *PoolService) *RemoveServiceFromPoolEndpoint {
	e := &RemoveServiceFromPoolEndpoint{}
	e.Construct(s, pool_api.RemoveServiceFromPool())
	return e
}
