package bare_bones_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_service"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Server interface {
	ApiServer() api_server.Server
	Auth() auth.Auth
	SmsManager() sms.SmsManager
}

type Config struct {
	Auth         auth.Auth
	Server       api_server.Server
	SmsManager   sms.SmsManager
	SmsProviders sms.ProviderFactory
}

type pimpl struct {
	auth         auth.Auth
	server       api_server.Server
	smsManager   sms.SmsManager
	smsProviders sms.ProviderFactory
	users        auth_session.WithUserSessionManager
}

type BareBonesServerBase struct {
	pimpl
}

func (s *BareBonesServerBase) Construct(users auth_session.WithUserSessionManager, config ...Config) {
	s.pimpl.users = users
	if len(config) != 0 {
		cfg := config[0]
		s.pimpl.server = cfg.Server
		s.pimpl.auth = cfg.Auth
		s.pimpl.smsManager = cfg.SmsManager
		s.pimpl.smsProviders = cfg.SmsProviders
	}
}

func New(users auth_session.WithUserSessionManager, config ...Config) *BareBonesServerBase {
	s := &BareBonesServerBase{}
	s.Construct(users, config...)
	return s
}

func (s *BareBonesServerBase) Init(app app_context.Context, tenancyManager multitenancy.Multitenancy, configPath ...string) error {

	path := utils.OptionalArg("server", configPath...)

	// init SMS manager
	if s.pimpl.smsManager == nil && s.pimpl.smsProviders != nil {
		smsManager := sms.NewSmsManager()
		err := smsManager.Init(app.Cfg(), app.Logger(), app.Validator(), s.pimpl.smsProviders, "sms")
		if err != nil {
			return app.Logger().PushFatalStack("failed to init SMS manager", err)
		}
		s.pimpl.smsManager = smsManager
	}

	// init auth controller
	if s.pimpl.auth == nil {
		auth := auth.NewAuth()
		authPath := object_config.Key(path, "auth")
		err := auth.Init(app.Cfg(), app.Logger(), app.Validator(), &auth_factory.DefaultAuthFactory{Users: s.pimpl.users, SmsManager: s.pimpl.smsManager}, authPath)
		if err != nil {
			return app.Logger().PushFatalStack("failed to init auth manager", err)
		}
		s.pimpl.auth = auth
	}

	// init REST API server
	if s.pimpl.server == nil {
		server := rest_api_gin_server.NewServer(tenancyManager)
		serverPath := object_config.Key(path, "rest_api_server")
		err := server.Init(app, s.pimpl.auth, serverPath)
		if err != nil {
			return app.Logger().PushFatalStack("failed to init REST API server", err)
		}
		s.pimpl.server = server
	}

	// add services
	api_server.AddServiceToServer(s.pimpl.server, api_server.NewStatusService())
	api_server.AddServiceToServer(s.pimpl.server, auth_service.NewAuthService())

	// done
	return nil
}

func (s *BareBonesServerBase) Auth() auth.Auth {
	return s.pimpl.auth
}

func (s *BareBonesServerBase) ApiServer() api_server.Server {
	return s.pimpl.server
}

func (s *BareBonesServerBase) SmsManager() sms.SmsManager {
	return s.pimpl.smsManager
}
