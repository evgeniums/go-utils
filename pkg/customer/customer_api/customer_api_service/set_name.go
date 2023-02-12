package customer_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/customer/customer_api"
)

type SetNameEndpoint struct {
	CustomerEndpoint
}

func (s *SetNameEndpoint) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("customer.SetName")
	defer request.TraceOutMethod()

	cmd := &common.WithNameBase{}
	err := request.ParseVerify(cmd)
	if err != nil {
		return err
	}

	err = Setter(s.service.Customers, request).SetName(request, request.GetResourceId("customer"), cmd.Name())
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func SetName(service *CustomerService) api_server.ResourceEndpointI {
	e := &SetNameEndpoint{}
	return e.Init(e, "name", service, customer_api.SetName())
}
