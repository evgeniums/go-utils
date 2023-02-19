package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type TenancyClient struct {
	api_client.ServiceClient

	TenanciesResource api.Resource
	TenancyResource   api.Resource

	add    api.Operation
	list   api.Operation
	exists api.Operation
}

func NewTenancyClient(client api_client.Client) *TenancyClient {

	c := &TenancyClient{}

	c.Init(client, tenancy_api.ServiceName)
	c.TenancyResource = api.NamedResource(tenancy_api.TenancyResource)
	c.TenanciesResource = c.TenancyResource.Parent()
	c.AddChild(c.TenanciesResource)

	c.add = tenancy_api.Add()
	c.list = tenancy_api.List()
	c.TenanciesResource.AddOperations(c.add,
		c.list,
	)

	existsResource := api.NewResource("exists")
	c.TenanciesResource.AddChild(existsResource)
	c.exists = tenancy_api.Exists()
	existsResource.AddOperation(c.exists)

	return c
}
