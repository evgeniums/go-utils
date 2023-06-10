package pool_misrocervice_client

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/app_with_pools"
)

type PoolServiceClient interface {
	api_client.Client
	InitForPoolService(service *pool.PoolServiceBinding, clientAgent ...string) error
}

type PoolMicroserviceClientConfig struct {
	POOL_SERVICE_ROLE string `validate:"required" vmessage:"Service role in the pool must be specified"`
}

type PoolMicroserviceClient struct {
	PoolMicroserviceClientConfig
	PoolServiceClient
}

func NewPoolMicroserviceClient(defaultRole string, client ...PoolServiceClient) *PoolMicroserviceClient {
	p := &PoolMicroserviceClient{}
	if len(client) != 0 {
		p.PoolServiceClient = client[0]
	} else {
		p.PoolServiceClient = &RestApiPoolServiceClient{}
	}
	p.POOL_SERVICE_ROLE = defaultRole
	return p
}

func (p *PoolMicroserviceClient) Config() interface{} {
	return &p.PoolMicroserviceClientConfig
}

func AppUserAgent(app app_context.Context) string {
	userAgent := fmt.Sprintf("%s/%s/%s", app.Hostname(), app.Application(), app.AppInstance())
	return userAgent
}

func (p *PoolMicroserviceClient) Init(app app_with_pools.AppWithPools, configPath ...string) error {

	// load config
	err := object_config.LoadLogValidate(app.Cfg(), app.Logger(), app.Validator(), p, "microservice_client", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load configuration of microservice api client", err)
	}

	// check if app with self pool
	selfPool, err := app.Pools().SelfPool()
	if err != nil {
		return app.Logger().PushFatalStack("self pool must be specified for microservice api client", err)
	}

	// find service for role
	service, err := selfPool.Service(p.POOL_SERVICE_ROLE)
	if err != nil {
		return app.Logger().PushFatalStack("failed to find service with specified role", err, logger.Fields{"role": p.POOL_SERVICE_ROLE})
	}

	// init client form service data
	err = p.InitForPoolService(service, AppUserAgent(app))
	if err != nil {
		return app.Logger().PushFatalStack("failed to init microservice api client with pool service configuration", err)
	}

	p.SetPropagateAuthUser(true)
	p.SetPropagateContextId(true)

	// done
	return nil
}

func (p *PoolMicroserviceClient) SetService(ctx op_context.Context, service *pool.PoolServiceBinding) error {

	c := ctx.TraceInMethod("PoolMicroserviceClient.SetService")
	defer ctx.TraceOutMethod()

	// init client form service data
	err := p.InitForPoolService(service, AppUserAgent(ctx.App()))
	if err != nil {
		c.SetMessage("failed to init microservice api client with pool service configuration")
		return c.SetError(err)
	}

	// done
	return nil
}

func (p *PoolMicroserviceClient) SetPropagateAuthUser(val bool) {
	p.PoolServiceClient.SetPropagateAuthUser(val)
}

func (p *PoolMicroserviceClient) SetPropagateContextId(val bool) {
	p.PoolServiceClient.SetPropagateContextId(val)
}
