package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
)

type ListPoolsEndpoint struct {
	PoolEndpoint
}

func (e *ListPoolsEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.ListPools")
	defer request.TraceOutMethod()

	// parse query
	queryName := request.Endpoint().Resource().ServicePathPrototype()
	filter, err := api_server.ParseDbQuery(request, &pool.PoolBase{}, queryName)
	if err != nil {
		return c.SetError(err)
	}

	// get services
	resp := &pool_api.ListPoolsResponse{}
	resp.Items, resp.Count, err = e.service.Pools.GetPools(request, filter)
	if err != nil {
		return c.SetError(err)
	}

	// TODO make support hateoas
	// if request.Server().IsHateoas() {
	// 	api.ProcessListResourceHateousLinks(request.Endpoint().Resource(), "service", resp.Services)
	// }

	// set response message
	request.Response().SetMessage(resp)

	// done
	return nil
}

func ListPools(s *PoolService) *ListPoolsEndpoint {
	e := &ListPoolsEndpoint{}
	e.Construct(s, pool_api.ListPools())
	return e
}
