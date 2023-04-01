package pool_misrocervice_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type RestApiPoolServiceClient struct {
	*rest_api_client.Client
}

func (r *RestApiPoolServiceClient) InitForPoolService(service *pool.PoolServiceBinding, clientAgent ...string) error {
	r.Client = rest_api_client.New(rest_api_client.DefaultRestApiClient(service.PRIVATE_URL, clientAgent...))
	return nil
}
