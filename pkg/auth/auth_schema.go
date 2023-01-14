package auth

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Aggregation = string

const (
	And Aggregation = "and"
	Or  Aggregation = "or"
)

type AuthSchemaConfig struct {
	NAME        string
	AGGREGATION string `default:"and" validate:"omitempty,oneof=and or"`
}

type AuthSchema struct {
	AuthHandlerBase
	config      *AuthSchemaConfig
	aggregation Aggregation
	handlers    []AuthHandler
}

func NewAuthSchema() *AuthSchema {
	s := &AuthSchema{}
	s.config = &AuthSchemaConfig{}
	s.handlers = make([]AuthHandler, 0)
	return s
}

func (a *AuthSchema) Config() interface{} {
	return a.config
}

func (a *AuthSchema) Name() string {
	return a.config.NAME
}

func (a *AuthSchema) Aggregation() Aggregation {
	return a.config.AGGREGATION
}

func (a *AuthSchema) Handlers() []AuthHandler {
	return a.handlers
}

func (a *AuthSchema) Init(log logger.Logger, cfg config.Config, vld validator.Validator, configPath ...string) error {
	return errors.New("use AuthSchema.InitSchema for initialization")
}

func (a *AuthSchema) InitSchema(log logger.Logger, cfg config.Config, vld validator.Validator, handlerStore HandlerStore, configPath ...string) error {

	path := utils.OptionalArg("auth_schema", configPath...)
	fields := logger.Fields{"path": path}

	// load plain configuration
	err := object_config.LoadLogValidate(cfg, log, vld, a, path)
	if err != nil {
		return log.Fatal("failed to load configuration of authorization schema", err, fields)
	}

	// load handlers
	key := object_config.Key(path, "handlers")
	handlersSection := cfg.Get(key)
	handlers := handlersSection.([]interface{})
	for i := range handlers {
		handlerPath := object_config.Key(key, fmt.Sprintf("%d", i))
		fields := logger.Fields{"path": handlerPath}

		handlerName := cfg.GetString(object_config.Key(handlerPath, "name"))
		if handlerName != "" {
			fields["handler"] = handlerName
			handler, err := handlerStore.Handler(handlerName)
			if err != nil {
				return log.Fatal("failed to find authorization handler in handlers store", err, fields)
			}
			a.handlers = append(a.handlers, handler)
		} else {
			childSchema := &AuthSchema{}
			err = childSchema.InitSchema(log, cfg, vld, handlerStore, handlerPath)
			if err != nil {
				return err
			}
			a.handlers = append(a.handlers, childSchema)
			if childSchema.Name() != "" {
				handlerStore.AddHandler(childSchema)
			}
		}
	}

	// done
	return nil
}

func (a *AuthSchema) Handle(ctx AuthContext, paramsResolver AuthParameterResolver) error {

	c := ctx.TraceInMethod("AuthSchema.Handle")
	defer ctx.TraceOutMethod()

	for _, handler := range a.handlers {
		err := handler.Handle(ctx, paramsResolver)
		if err != nil {
			ctx.SetGenericError(generic_error.New(ErrorCodeUnauthorized))
			return c.Check(err)
		}
		if a.aggregation == Or {
			return nil
		}
	}
	return nil
}
