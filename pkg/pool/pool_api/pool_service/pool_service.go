package pool_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolEndpoint struct {
	service *PoolService
	api_server.EndpointBase
}

func (e *PoolEndpoint) Construct(service *PoolService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type PoolService struct {
	api_server.ServiceBase
	Pools pool.PoolController

	PoolsResource    api.Resource
	PoolResource     api.Resource
	ServicesResource api.Resource
	ServiceResource  api.Resource
}

func NewPoolService(poolController pool.PoolController) *PoolService {

	s := &PoolService{}
	s.Pools = poolController

	var serviceName string
	serviceName, s.PoolsResource, s.PoolResource = api.PrepareCollectionAndNameResource("pool")
	s.Init(serviceName)
	s.AddChild(s.PoolsResource)

	_, s.ServicesResource, s.ServiceResource = api.PrepareCollectionAndNameResource("service")
	s.AddChild(s.ServicesResource)

	s.PoolsResource.AddOperation(AddPool(s))

	return s
}
