package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ListServicesEndpoint struct {
	PoolEndpoint
}

func (e *ListServicesEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("pool.FindService")
	defer request.TraceOutMethod()

	// parse query
	cmd := &api.DbQuery{}
	queryName := request.Endpoint().Resource().ServicePathPrototype()
	filter, err := api_server.ParseDbQuery(request, utils.List(&pool.PoolServiceBase{}), cmd, queryName)
	if err != nil {
		return c.SetError(err)
	}

	// get services
	resp := &pool_api.ListServicesResponse{}
	resp.Services, resp.Count, err = e.service.Pools.GetServices(request, filter)
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

func ListServices(s *PoolService) *ListServicesEndpoint {
	e := &ListServicesEndpoint{}
	e.Construct(s, pool_api.ListServices())
	return e
}
