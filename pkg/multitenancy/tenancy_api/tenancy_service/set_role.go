package tenancy_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type SetRoleEndpoint struct {
	TenancyUpdateEndpoint
}

func (s *SetRoleEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("tenancy.SetRole")
	defer request.TraceOutMethod()

	// parse command
	cmd := &multitenancy.WithRole{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("faield to parse/validate command")
		return err
	}

	// apply
	err = s.service.Tenancies.SetRole(request, request.GetResourceId(tenancy_api.TenancyResource), cmd.Role())
	if err != nil {
		return c.SetError(err)
	}

	// done
	return nil
}

func SetRole(s *TenancyService) *SetRoleEndpoint {
	e := &SetRoleEndpoint{}
	e.Construct(s, e, "role", tenancy_api.SetRole())
	return e
}
