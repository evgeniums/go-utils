package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type AuthDataAccessor interface {
	Set(key string, value string)
	Get(key string) string
}

type AuthContext interface {
	UserContext
	Session

	GetRequestContent() []byte
	CheckRequestContent(smsMessage *string) error
	GetRequestPath() string
	GetRequestMethod() string

	GetRequestClientIp() string
	GetRequestUserAgent() string

	SetAuthParameter(authMethodProtocol string, key string, value string)
	GetAuthParameter(authMethodProtocol string, key string) string
}

type AuthHandler interface {
	common.WithName

	Protocol() string

	Handle(ctx AuthContext) (bool, error)
	Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error
	Handlers() []AuthHandler

	ErrorDescriptions() map[string]string
	ErrorProtocolCodes() map[string]int
}

type AuthHandlerBase struct {
	common.WithNameBase

	protocol string
}

func (a *AuthHandlerBase) Init(protocol string) {
	a.protocol = protocol
	a.WithNameBase.Init(protocol)
}

func (a *AuthHandlerBase) ErrorDescriptions() map[string]string {
	return map[string]string{}
}

func (a *AuthHandlerBase) ErrorProtocolCodes() map[string]int {
	return map[string]int{}
}

func (a *AuthHandlerBase) Protocol() string {
	return a.protocol
}

func (a *AuthHandlerBase) Handlers() []AuthHandler {
	return nil
}
