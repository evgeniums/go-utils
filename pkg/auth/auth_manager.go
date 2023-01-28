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
	Create(protocol string) (AuthHandler, error)
}

type AuthManager interface {
	Handle(ctx AuthContext, schema string) error
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

func (a *AuthManagerBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, handlerFactory HandlerFactory, configPath ...string) error {

	path := utils.OptionalArg("auth_manager", configPath...)
	fields := logger.Fields{"config_path": path}
	log.Info("Init authorization manager", fields)

	a.store = NewHandlerStore()

	// create and init auth methods
	log.Debug("Init auth methods")
	methodsPath := object_config.Key(path, "methods")
	methodsSection := cfg.Get(methodsPath)
	methods, ok := methodsSection.(map[string]interface{})
	if !ok {
		return log.PushFatalStack("failed to initialize authorization methods", errors.New("invalid methods section"), fields)
	}
	for methodProtocol := range methods {
		methodPath := object_config.Key(methodsPath, methodProtocol)
		fields := utils.AppendMapNew(fields, logger.Fields{"auth_method": methodProtocol, "config_path": methodPath})
		log.Debug("Init auth method", fields)
		handler, err := handlerFactory.Create(methodProtocol)
		if err != nil {
			return log.PushFatalStack("failed to create authorization method", err, fields)
		}
		err = handler.Init(cfg, log, vld, methodPath)
		if err != nil {
			return log.PushFatalStack("failed to initialize authorization method", err, fields)
		}
		a.store.AddHandler(handler)
	}

	// create and init auth schemas
	log.Debug("Init auth schemas")
	schemasPath := object_config.Key(path, "schemas")
	if cfg.IsSet("schemas") {
		schemasSection := cfg.Get(schemasPath)
		schemas, ok := schemasSection.([]interface{})
		if !ok {
			return log.PushFatalStack("failed to initialize authorization schemas", errors.New("invalid schemas section"), fields)
		}
		for i := range schemas {
			schemaPath := object_config.KeyInt(path, i)
			fields := utils.AppendMapNew(fields, logger.Fields{"schema_path": schemaPath})
			log.Debug("Init auth schema", fields)
			schema := NewAuthSchema()
			err := schema.InitSchema(log, cfg, vld, a.store, schemaPath)
			if err != nil {
				return log.PushFatalStack("failed to initialize authorization schema", err, fields)
			}
			a.store.AddHandler(schema)
			for _, subhandler := range schema.Handlers() {
				a.store.AddHandler(subhandler)
			}
		}
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

func (a *AuthManagerBase) Handle(ctx AuthContext, schema string) error {

	// setup
	c := ctx.TraceInMethod("AuthManagerBase.Handle", logger.Fields{"schema": schema})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find handler
	handler, err := a.store.Handler(schema)
	if err != nil {
		ctx.SetGenericErrorCode(ErrorCodeInvalidAuthSchema)
		return c.SetError(err)
	}

	// run handler
	_, err = handler.Handle(ctx)
	return err
}
