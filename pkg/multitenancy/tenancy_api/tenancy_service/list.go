package tenancy_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type ListEndpoint struct {
	TenancyEndpoint
}

func (e *ListEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("tenancy.List")
	defer request.TraceOutMethod()

	// parse query
	queryName := request.Endpoint().Resource().ServicePathPrototype()
	filter, err := api_server.ParseDbQuery(request, &multitenancy.TenancyDb{}, queryName)
	if err != nil {
		return c.SetError(err)
	}

	// get
	resp := &tenancy_api.ListTenanciesResponse{}
	resp.Tenancies, resp.Count, err = e.service.Tenancies.List(request, filter)
	if err != nil {
		return c.SetError(err)
	}

	// TODO make support hateoas
	// if request.Server().IsHateoas() {
	// }

	// set response message
	request.Response().SetMessage(resp)

	// done
	return nil
}

func List(s *TenancyService) *ListEndpoint {
	e := &ListEndpoint{}
	e.Construct(s, tenancy_api.List())
	return e
}
