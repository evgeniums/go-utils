package tenancy_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type AddEndpoint struct {
	TenancyEndpoint
}

func (e *AddEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("tenancy.Add")
	defer request.TraceOutMethod()

	// parse command
	cmd := &multitenancy.TenancyData{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("faield to parse/validate command")
		return err
	}

	// add
	resp := &tenancy_api.TenancyResponse{}
	resp.TenancyItem, err = e.service.Tenancies.Add(request, cmd)
	if err != nil {
		c.SetMessage("failed to add tenancy")
		return c.SetError(err)
	}

	// set response
	request.Response().SetMessage(resp)

	// done
	return nil
}

func Add(s *TenancyService) *AddEndpoint {
	e := &AddEndpoint{}
	e.Construct(s, tenancy_api.Add())
	return e
}
