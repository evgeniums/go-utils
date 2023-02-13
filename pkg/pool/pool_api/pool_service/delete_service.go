package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
)

type DeleteServiceEndpoint struct {
	PoolEndpoint
}

func (e *DeleteServiceEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.DeleteService")
	defer request.TraceOutMethod()

	// delete pool
	err := e.service.Pools.DeleteService(request, request.GetResourceId("service"))
	if err != nil {
		c.SetMessage("failed to delete service")
		return c.SetError(err)
	}

	// done
	return nil
}

func DeleteService(s *PoolService) *DeleteServiceEndpoint {
	e := &DeleteServiceEndpoint{}
	e.Construct(s, api.Delete())
	return e
}
