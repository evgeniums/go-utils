package pool_microservice_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/noauth_server"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
)

type PoolMicroserviceServerConfig struct {
	POOL_SERVICE_NAME string
	POOL_SERVICE_TYPE string
}

type PoolMicroserviceServer struct {
	config PoolMicroserviceServerConfig
	noauth_server.NoAuthServer
}

type Config struct {
	noauth_server.Config
}

func New(defaultPoolServiceType string, config ...*Config) *PoolMicroserviceServer {
	s := &PoolMicroserviceServer{}
	s.Construct(config...)
	s.config.POOL_SERVICE_TYPE = defaultPoolServiceType
	return s
}

func (s *PoolMicroserviceServer) Config() interface{} {
	return &s.config
}

func (s *PoolMicroserviceServer) Construct(config ...*Config) {
	if len(config) != 0 {
		s.NoAuthServer.Construct(config[0].Config)
	} else {
		s.NoAuthServer.Construct()
	}
}

func (s *PoolMicroserviceServer) Init(app app_with_multitenancy.AppWithMultitenancy, configPath ...string) error {

	err := object_config.LoadLogValidate(app.Cfg(), app.Logger(), app.Validator(), s, "microservice_server", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load microservice server configuration", err)
	}

	if s.config.POOL_SERVICE_NAME != "" {

		// check if app with self pool
		selfPool, err := app.Pools().SelfPool()
		if err != nil {
			return app.Logger().PushFatalStack("self pool must be specified for microservice api server", err)
		}

		// find service for role
		service, err := selfPool.ServiceByName(s.config.POOL_SERVICE_NAME)
		if err != nil {
			return app.Logger().PushFatalStack("failed to find service with specified name", err, logger.Fields{"name": s.config.POOL_SERVICE_NAME})
		}

		if service.TypeName() != s.config.POOL_SERVICE_TYPE {
			return app.Logger().PushFatalStack("invalid service type", err, logger.Fields{"name": s.config.POOL_SERVICE_NAME, "service_type": s.config.POOL_SERVICE_TYPE, "pool_service_type": service.TypeName()})
		}

		if service.Provider() != app.Application() {
			return app.Logger().PushFatalStack("invalid service type", err, logger.Fields{"name": s.config.POOL_SERVICE_NAME, "service_type": s.config.POOL_SERVICE_TYPE, "pool_service_type": service.TypeName()})
		}

		// load server configuration from service
		s.NoAuthServer.SetConfigFromPoolService(service)
	}

	return s.NoAuthServer.Init(app, configPath...)
}
