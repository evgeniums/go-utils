package pool_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api"
)

type PoolClient struct {
	api_client.ServiceClient

	PoolsResource    api.Resource
	PoolResource     api.Resource
	ServicesResource api.Resource
	ServiceResource  api.Resource

	add_pool    api.Operation
	add_service api.Operation
}

func NewPoolClient(client api_client.Client) *PoolClient {

	c := &PoolClient{}
	var serviceName string
	serviceName, c.PoolsResource, c.PoolResource = api.PrepareCollectionAndNameResource("pool")
	c.Init(client, serviceName)
	c.AddChild(c.PoolsResource)

	_, c.ServicesResource, c.ServiceResource = api.PrepareCollectionAndNameResource("service")
	c.AddChild(c.ServicesResource)

	c.add_pool = pool_api.AddPool()
	c.PoolsResource.AddOperations(c.add_pool)

	c.add_service = pool_api.AddService()
	c.ServicesResource.AddOperations(c.add_service)

	return c
}
