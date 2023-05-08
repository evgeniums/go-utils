package pool_misrocervice_client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type RestApiPoolServiceClient struct {
	*rest_api_client.Client
}

func (r *RestApiPoolServiceClient) InitForPoolService(service *pool.PoolServiceBinding, clientAgent ...string) error {

	url := service.PrivateUrl()
	if url == "" {
		host := service.PrivateHost()
		if host == "" {
			return errors.New("unknown URL of the service")
		}
		if !strings.HasPrefix(host, "http") {
			host = utils.ConcatStrings("http://", host)
		}
		if service.PrivatePort() == 0 {
			return errors.New("unknown port of the service")
		}
		apiVersion := ""
		if service.ApiVersion() != "" {
			apiVersion = fmt.Sprintf("/%s", service.ApiVersion())
		}
		url = fmt.Sprintf("%s:%d%s%s", host, service.PrivatePort(), service.PathPrefix(), apiVersion)
	}

	r.Client = rest_api_client.New(rest_api_client.DefaultRestApiClient(url, clientAgent...))
	return nil
}
