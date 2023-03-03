package tenancy_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type TenancyEndpoint struct {
	service *TenancyService
	api_server.EndpointBase
}

func (e *TenancyEndpoint) Construct(service *TenancyService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type TenancyUpdateEndpoint struct {
	service *TenancyService
	api_server.ResourceEndpoint
}

func (e *TenancyUpdateEndpoint) Construct(service *TenancyService, ep api_server.ResourceEndpointI, epName string, op api.Operation) {
	e.service = service
	api_server.ConstructResourceEndpoint(ep, epName, op)
}

type TenancyService struct {
	api_server.ServiceBase
	Tenancies multitenancy.TenancyController

	TenanciesResource api.Resource
	TenancyResource   api.Resource
}

func NewTenancyService(tenancyController multitenancy.TenancyController) *TenancyService {

	s := &TenancyService{}
	s.ErrorsExtenderBase.Init(multitenancy.ErrorDescriptions, multitenancy.ErrorHttpCodes)
	s.ErrorsExtenderBase.AddErrors(customer.ErrorDescriptions, customer.ErrorHttpCodes)
	s.Tenancies = tenancyController

	s.Init(tenancy_api.ServiceName)
	s.TenancyResource = api.NamedResource(tenancy_api.TenancyResource)
	s.TenanciesResource = s.TenancyResource.Parent()
	s.AddChild(s.TenanciesResource)

	listOp := List(s)
	s.TenanciesResource.AddOperations(Add(s), listOp)
	existsResource := api.NewResource("exists")
	existsResource.AddOperation(Exists(s))
	s.TenanciesResource.AddChild(existsResource)

	s.TenancyResource.AddOperation(Find(s), true)
	s.TenancyResource.AddOperation(Delete(s))
	s.TenancyResource.AddChildren(SetActive(s),
		SetPath(s),
		SetRole(s),
		SetCustomer(s),
		ChangePoolOrDb(s),
	)

	tenancyTableConfig := &api_server.DynamicTableConfig{Model: &multitenancy.TenancyItem{}, Operation: listOp}
	s.AddDynamicTables(tenancyTableConfig)

	return s
}
