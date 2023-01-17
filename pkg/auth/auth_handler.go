package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type User interface {
	common.Object
	Display() string
	GetAuthParameter(methodProtocol string, key string) string
}

type AuthDataAccessor interface {
	Set(key string, value string)
	Get(key string) string
}

type AuthContext interface {
	op_context.Context

	GetRequestContent() []byte
	CheckRequestContent(authDataAccessor ...AuthDataAccessor) error

	AuthUser() User

	SetAuthParameter(authMethodProtocol string, key string, value string)
	GetAuthParameter(authMethodProtocol string, key string) string
}

type AuthHandler interface {
	common.WithName

	Protocol() string

	Handle(ctx AuthContext) (bool, error)
	Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error

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
