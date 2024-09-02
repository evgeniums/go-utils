package pool_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/pool/pool_api"
)

type ListServicePoolsEndpoint struct {
	PoolEndpoint
}

func (e *ListServicePoolsEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	var err error
	c := request.TraceInMethod("pool.ListServicePools")
	defer request.TraceOutMethod()

	// find service
	resp := &pool_api.ListServicePoolsResponse{}
	resp.Items, err = e.service.Pools.GetServiceBindings(request, request.GetResourceId("service"))
	if err != nil {
		c.SetMessage("failed to get service bindings")
		return c.SetError(err)
	}

	// set response
	request.Response().SetMessage(resp)

	// done
	return nil
}

func ListServicePools(s *PoolService) *ListServicePoolsEndpoint {
	e := &ListServicePoolsEndpoint{}
	e.Construct(s, pool_api.ListServicePools())
	return e
}
