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

type AuthBase struct {
	manager         AuthManager
	endpointsConfig EndpointsAuthConfig
}

func (a *AuthBase) Init(log logger.Logger, cfg config.Config, vld validator.Validator, handlerFactory HandlerFactory, configPath ...string) error {

	path := utils.OptionalArg("auth", configPath...)

	manager := &AuthManagerBase{}
	err := manager.Init(log, cfg, vld, handlerFactory, object_config.Key(path, "manager"))
	if err != nil {
		return err
	}
	a.manager = manager

	epConfig := &EndpointsAuthConfigBase{}
	err = epConfig.Init(log, cfg, vld, object_config.Key(path, "endpoints_config"))
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
		schema = a.endpointsConfig.DefaultSchema()
	}

	return a.manager.Handle(ctx, schema)
}

func (a *AuthBase) AttachToErrorManager(errManager generic_error.ErrorManager) {
	errManager.AddErrorDescriptions(a.manager.ErrorDescriptions())
	errManager.AddErrorProtocolCodes(a.manager.ErrorProtocolCodes())
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
