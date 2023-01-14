package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

// Accessor request's parameters used in auth handler.
type AuthParameters interface {
	SetAuthParameter(key string, value string)
	GetAuthParameter(key string) string
	GetRequestContent() []byte
}

// Type used to resolve AuthParameters for specific auth method.
type AuthParameterResolver = func(methodProtocol string) AuthParameters

type User interface {
	common.Object
	Display() string
	GetAuthParameter(key string) string
}

type AuthContext interface {
	op_context.Context
	AuthUser() User
}

type AuthHandler interface {
	common.WithName
	Handle(ctx AuthContext, paramsResolver AuthParameterResolver) error
	Init(log logger.Logger, cfg config.Config, vld validator.Validator, configPath ...string) error

	ErrorDescriptions() map[string]string
	ErrorProtocolCodes() map[string]int
}

type AuthHandlerBase struct{}

func (a *AuthHandlerBase) ErrorDescriptions() map[string]string {
	return map[string]string{}
}

func (a *AuthHandlerBase) ErrorProtocolCodes() map[string]int {
	return map[string]int{}
}
