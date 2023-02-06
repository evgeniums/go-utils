package auth_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_service"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type AuthServer interface {
	SmsManager() sms.SmsManager
	ApiServer() api_server.Server
	Auth() auth.Auth
}

type AuthServerBase struct {
	SmsManagerBase *sms.SmsManagerBase
	AuthBase       *auth.AuthBase
	RestApiServer  *rest_api_gin_server.Server
}

func (s *AuthServerBase) Construct() {
	s.SmsManagerBase = sms.NewSmsManager()
	s.RestApiServer = rest_api_gin_server.NewServer()
	s.AuthBase = auth.NewAuth()
}

func NewAuthServer() *AuthServerBase {
	s := &AuthServerBase{}
	s.Construct()
	return s
}

func (s *AuthServerBase) Init(app app_context.Context, users user_manager.WithSessionManager, smsProviders sms.ProviderFactory, configPath ...string) error {

	path := utils.OptionalArg("auth_server", configPath...)

	// init SMS manager
	err := s.SmsManagerBase.Init(app.Cfg(), app.Logger(), app.Validator(), smsProviders, "sms")
	if err != nil {
		return app.Logger().PushFatalStack("failed to init SMS manager", err)
	}

	// init auth controller
	authPath := object_config.Key(path, "auth")
	err = s.AuthBase.Init(app.Cfg(), app.Logger(), app.Validator(), &auth_factory.DefaultAuthFactory{Users: users, SmsManager: s.SmsManagerBase}, authPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init auth manager", err)
	}

	// init REST API server
	serverPath := object_config.Key(path, "rest_api_server")
	err = s.RestApiServer.Init(app, s.AuthBase, serverPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init REST API server", err)
	}
	s.SmsManagerBase.AttachToErrorManager(s.RestApiServer)

	// add services
	api_server.AddServiceToServer(s.RestApiServer, api_server.NewStatusService())
	api_server.AddServiceToServer(s.RestApiServer, auth_service.NewAuthService())

	// done
	return nil
}

func (s *AuthServerBase) SmsManager() sms.SmsManager {
	return s.SmsManagerBase
}

func (s *AuthServerBase) Auth() auth.Auth {
	return s.AuthBase
}

func (s *AuthServerBase) ApiServer() api_server.Server {
	return s.RestApiServer
}
