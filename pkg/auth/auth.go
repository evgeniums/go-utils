package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Auth interface {
	generic_error.ErrorDefinitions

	Manager() AuthManager
	EndpointsConfig() EndpointsAuthConfig

	HandleRequest(ctx AuthContext, path string, access access_control.AccessType) error
}

type AuthBaseConfig struct {
	DEFAULT_SCHEMA string `default:"token"`
}

type AuthBase struct {
	AuthBaseConfig
	manager         AuthManager
	endpointsConfig EndpointsAuthConfig
}

func (a *AuthBase) Config() interface{} {
	return &a.AuthBaseConfig
}

func (a *AuthBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, handlerFactory HandlerFactory, configPath ...string) error {

	path := utils.OptionalArg("auth", configPath...)

	err := object_config.LoadLogValidate(cfg, log, vld, a, path)
	if err != nil {
		return log.Fatal("Failed to load auth configuration", err)
	}

	manager := &AuthManagerBase{}
	err = manager.Init(cfg, log, vld, handlerFactory, object_config.Key(path, "manager"))
	if err != nil {
		return err
	}
	a.manager = manager

	epConfig := &EndpointsAuthConfigBase{}
	err = epConfig.Init(cfg, log, vld, object_config.Key(path, "endpoints"))
	if err != nil {
		return err
	}
	a.endpointsConfig = epConfig

	return nil
}

func (a *AuthBase) HandleRequest(ctx AuthContext, path string, access access_control.AccessType) error {

	ctx.TraceInMethod("AuthBase.HandleRequest")
	defer ctx.TraceOutMethod()

	schema, ok := a.endpointsConfig.Schema(path, access)
	if !ok {
		schema = a.DEFAULT_SCHEMA
	}

	return a.manager.Handle(ctx, schema)
}

func (a *AuthBase) AttachToErrorManager(errManager generic_error.ErrorManager) {
	errManager.AddErrorDescriptions(a.manager.ErrorDescriptions())
	errManager.AddErrorProtocolCodes(a.manager.ErrorProtocolCodes())
}

func (a *AuthBase) EndpointsConfig() EndpointsAuthConfig {
	return a.endpointsConfig
}

func (a *AuthBase) Manager() AuthManager {
	return a.manager
}

type WithAuth interface {
	Auth() Auth
}

type WithAuthBase struct {
	auth Auth
}

func (w *WithAuthBase) Init(auth Auth) {
	w.auth = auth
}

func (w *WithAuthBase) Auth() Auth {
	return w.auth
}
