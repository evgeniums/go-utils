package auth

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type HandlerStore interface {
	Handler(name string) (AuthHandler, error)
	AddHandler(handler AuthHandler)
	HandlerNames() []string
}

type HandlerFactory interface {
	Create(name string) (AuthHandler, error)
}

type AuthManager interface {
	Handle(ctx AuthContext, schema string, paramsResolver AuthParameterResolver) error
	Store() HandlerStore
	ErrorDescriptions() map[string]string
	ErrorProtocolCodes() map[string]int
}

type HandlerStoreBase struct {
	handlers map[string]AuthHandler
}

func NewHandlerStore() *HandlerStoreBase {
	s := &HandlerStoreBase{}
	s.handlers = make(map[string]AuthHandler)
	return s
}

func (h *HandlerStoreBase) Handler(name string) (AuthHandler, error) {
	handler, ok := h.handlers[name]
	if !ok {
		return nil, errors.New("unknown authorization handler")
	}
	return handler, nil
}

func (h *HandlerStoreBase) HandlerNames() []string {
	return utils.AllMapKeys(h.handlers)
}

func (h *HandlerStoreBase) AddHandler(handler AuthHandler) {
	h.handlers[handler.Name()] = handler
}

type AuthManagerBase struct {
	store HandlerStore
}

func (a *AuthManagerBase) Init(log logger.Logger, cfg config.Config, vld validator.Validator, handlerFactory HandlerFactory, configPath ...string) error {

	path := utils.OptionalArg("auth_manager", configPath...)
	fields := logger.Fields{"where": "AuthManagerBase.Init", "config_path": path}
	log.Info("Init authorization manager", fields)

	a.store = NewHandlerStore()

	// create and init auth methods
	methodsPath := object_config.Key(path, "methods")
	methodsSection := cfg.Get(methodsPath)
	methods := methodsSection.(map[string]interface{})
	for methodName := range methods {
		methodPath := object_config.Key(path, methodName)
		fields := utils.AppendMapNew(fields, logger.Fields{"method": methodName, "method_path": methodPath})
		handler, err := handlerFactory.Create(methodName)
		if err != nil {
			return log.Fatal("failed to create authorization method", err, fields)
		}
		err = handler.Init(log, cfg, vld, methodPath)
		if err != nil {
			return log.Fatal("failed to initialize authorization method", err, fields)
		}
		a.store.AddHandler(handler)
	}

	// create and init auth schemas
	schemasPath := object_config.Key(path, "schemas")
	schemasSection := cfg.Get(schemasPath)
	schemas := schemasSection.([]interface{})
	for i := range schemas {
		schemaPath := object_config.KeyInt(path, i)
		fields := utils.AppendMapNew(fields, logger.Fields{"schema_path": schemaPath})
		schema := NewAuthSchema()
		err := schema.InitSchema(log, cfg, vld, a.store, schemaPath)
		if err != nil {
			return log.Fatal("failed to initialize authorization schema", err, fields)
		}
		a.store.AddHandler(schema)
	}

	// done
	return nil
}

func (a *AuthManagerBase) Store() HandlerStore {
	return a.store
}

func (a *AuthManagerBase) ErrorDescriptions() map[string]string {
	m := utils.CopyMapOneLevel(ErrorDescriptions)
	for _, name := range a.store.HandlerNames() {
		handler, _ := a.store.Handler(name)
		utils.AppendMap(m, handler.ErrorDescriptions())
	}
	return m
}

func (a *AuthManagerBase) ErrorProtocolCodes() map[string]int {
	m := utils.CopyMapOneLevel(ErrorHttpCodes)
	for _, name := range a.store.HandlerNames() {
		handler, _ := a.store.Handler(name)
		utils.AppendMap(m, handler.ErrorProtocolCodes())
	}
	return m
}

func (a *AuthManagerBase) Handle(ctx AuthContext, schema string, paramsResolver AuthParameterResolver) error {
	c := ctx.TraceInMethod("AuthManagerBase.Handle", logger.Fields{"schema": schema})
	defer ctx.TraceOutMethod()

	handler, err := a.store.Handler(schema)
	if err != nil {
		ctx.SetGenericError(ctx.MakeGenericError(ErrorCodeInvalidAuthSchema))
		return c.Check(err)
	}
	return handler.Handle(ctx, paramsResolver)
}
