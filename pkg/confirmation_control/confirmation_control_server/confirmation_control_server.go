package confirmation_control_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	finish "github.com/evgeniums/go-finish-service"
)

type ConfirmationControlServer struct {
	*ExternalServer
	*InternalServer
}

func New(externalServerCfg ...ExternalServerCfg) *ConfirmationControlServer {
	s := &ConfirmationControlServer{}
	s.ExternalServer = NewExternalServer(externalServerCfg...)
	s.InternalServer = NewInternalServer()
	return s
}

func (s *ConfirmationControlServer) Init(app app_with_multitenancy.AppWithMultitenancy, configPath ...string) error {

	path := utils.OptionalArg("confirmation_control_server", configPath...)
	externalServerConfigPath := object_config.Key(path, "external_server")
	internalServerConfigPath := object_config.Key(path, "internal_server")

	// init external server
	err := s.ExternalServer.Init(app, externalServerConfigPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init external server", err)
	}

	// init internal server
	err = s.InternalServer.Init(app, internalServerConfigPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init internal server", err)
	}

	// done
	return nil
}

func (s *ConfirmationControlServer) Run(fin *finish.Finisher) {
	s.ExternalServer.ApiServer().Run(fin)
	s.InternalServer.ApiServer().Run(fin)
}
