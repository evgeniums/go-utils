package confirmation_control_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/pool_microservice/pool_microservice_server"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api/confirmation_api_service"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type InternalServerConfig struct {
	TOKEN_TTL int `default:"180" validate:"gte=1" vmessage:"Token TTL must be positive"`
}

const InternalServerType = "confirmation_control_internal"

type InternalServer struct {
	InternalServerConfig

	basePublicUrl string
	*pool_microservice_server.PoolMicroserviceServer
}

func NewInternalServer() *InternalServer {
	s := &InternalServer{}
	return s
}

func (s *InternalServer) Config() interface{} {
	return &s.InternalServerConfig
}

func (s *InternalServer) Init(app app_with_multitenancy.AppWithMultitenancy, basePublicUrl string, configPath ...string) error {

	s.basePublicUrl = basePublicUrl

	path := utils.OptionalArg("internal_server", configPath...)

	err := object_config.LoadLogValidate(app.Cfg(), app.Logger(), app.Validator(), s, path)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init internal server of confirmation control", err)
	}

	// init microservice server for internal requests
	s.PoolMicroserviceServer = pool_microservice_server.New(InternalServerType)
	err = s.PoolMicroserviceServer.Init(app, path)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init microservice server for internal server", err)
	}

	// create and add service
	service := confirmation_api_service.NewConfirmationInternalService(s.basePublicUrl, s.TOKEN_TTL)
	api_server.AddServiceToServer(s.PoolMicroserviceServer.ApiServer(), service)

	// done
	return nil
}
