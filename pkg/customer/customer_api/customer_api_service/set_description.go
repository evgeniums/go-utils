package customer_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/customer/customer_api"
)

type SetDescriptionEndpoint struct {
	CustomerEndpoint
}

func (s *SetDescriptionEndpoint) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("customer.SetDescription")
	defer request.TraceOutMethod()

	cmd := &common.WithDescriptionBase{}
	err := request.ParseValidate(cmd)
	if err != nil {
		return err
	}

	err = Setter(s.service.Customers, request).SetName(request, request.GetResourceId("customer"), cmd.Description())
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func SetDescription(service *CustomerService) api_server.ResourceEndpointI {
	e := &SetDescriptionEndpoint{}
	return e.Init(e, "description", service, customer_api.SetDescription())
}
