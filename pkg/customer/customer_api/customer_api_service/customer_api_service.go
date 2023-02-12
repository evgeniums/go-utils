package customer_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/customer/customer_api"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api/user_service"
)

type CustomerService struct {
	*user_service.UserService[*customer.Customer]
	Customers customer.CustomerController
}

func NewCustomerService(customers *customer.Manager) *CustomerService {
	c := &CustomerService{Customers: customers}
	c.UserService = user_service.NewUserService[*customer.Customer](customers, customer_api.NewCustomerFieldsSetter, "customer")

	c.UserService.UserResource().AddChildren(SetName(c), SetDescription(c))

	return c
}

type CustomerEndpoint struct {
	api_server.ResourceEndpoint
	service *CustomerService
}

func (e *CustomerEndpoint) Init(ep api_server.ResourceEndpointI, fieldName string, s *CustomerService, op api.Operation) api_server.ResourceEndpointI {
	api_server.ConstructResourceEndpoint(ep, fieldName, op)
	e.service = s
	return ep
}

type TenancyWithCustomer interface {
	CustomerFieldSetter() customer.CustomerFieldSetter
}

func Setter(setter customer.CustomerFieldSetter, request api_server.Request) customer.CustomerFieldSetter {

	t := request.GetTenancy()
	if t != nil {
		ts, ok := t.(TenancyWithCustomer)
		if ok {
			return ts.CustomerFieldSetter()
		}
	}

	return setter
}
