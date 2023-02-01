package auth

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Aggregation = string

const AggregationProtocol = "aggregation"

const (
	And Aggregation = "and"
	Or  Aggregation = "or"
)

type AuthSchemaConfig struct {
	common.WithNameBaseConfig
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
	s.Setup()
	return s
}

func (s *AuthSchema) Setup() {
	s.config = &AuthSchemaConfig{}
	s.handlers = make([]AuthHandler, 0)
}

func (a *AuthSchema) Config() interface{} {
	return a.config
}

func (a *AuthSchema) Aggregation() Aggregation {
	return a.config.AGGREGATION
}

func (a *AuthSchema) SetAggregation(aggregation Aggregation) {
	a.config.AGGREGATION = aggregation
}

func (a *AuthSchema) Handlers() []AuthHandler {
	return a.handlers
}

func (a *AuthSchema) AppendHandlers(handler ...AuthHandler) {
	a.handlers = append(a.handlers, handler...)
}

func (a *AuthSchema) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {
	return errors.New("use AuthSchema.InitSchema for initialization")
}

func (a *AuthSchema) InitSchema(log logger.Logger, cfg config.Config, vld validator.Validator, handlerStore HandlerStore, configPath ...string) error {

	path := utils.OptionalArg("auth_schema", configPath...)
	fields := logger.Fields{"path": path}

	// load plain configuration
	err := object_config.LoadLogValidate(cfg, log, vld, a, path)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of authorization schema", err, fields)
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
				return log.PushFatalStack("failed to find authorization handler in handlers store", err, fields)
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

func (a *AuthSchema) Handle(ctx AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthSchema.Handle", logger.Fields{"path": ctx.GetRequestPath()})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	authMethodFound := false
	for _, handler := range a.handlers {
		sectionFound, err := handler.Handle(ctx)
		if !handler.IsReal() || !sectionFound && a.aggregation == Or {
			continue
		}
		if err != nil {
			ctx.SetGenericError(ctx.MakeGenericError(ErrorCodeUnauthorized))
			return sectionFound, c.SetError(err)
		}
		if a.aggregation == Or {
			return sectionFound, nil
		}
		if sectionFound {
			authMethodFound = true
		}
	}
	if len(a.handlers) != 0 && !authMethodFound {
		err := errors.New("no auth section in request")
		ctx.SetGenericErrorCode(ErrorCodeUnauthorized)
		return authMethodFound, c.SetError(err)
	}
	return authMethodFound, nil
}

func (a *AuthSchema) Protocol() string {
	return AggregationProtocol
}
