package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
)

type DeletePoolEndpoint struct {
	PoolEndpoint
}

func (e *DeletePoolEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.DeletePool")
	defer request.TraceOutMethod()

	// delete pool
	err := e.service.Pools.DeletePool(request, request.GetResourceId("pool"))
	if err != nil {
		c.SetMessage("failed to delete pool")
		return c.SetError(err)
	}

	// done
	return nil
}

func DeletePool(s *PoolService) *DeletePoolEndpoint {
	e := &DeletePoolEndpoint{}
	e.Construct(s, api.Delete())
	return e
}
