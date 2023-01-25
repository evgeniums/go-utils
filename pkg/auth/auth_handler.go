package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type User interface {
	GetID() string
	Display() string
	Login() string
	IsBlocked() bool
}

type Session interface {
	GetSessionId() string
	SetSessionId(id string)
	GetClientId() string
	SetClientId(id string)
}

type SessionBase struct {
	session string
	client  string
}

func (u *SessionBase) GetSessionId() string {
	return u.session
}

func (u *SessionBase) SetSessionId(id string) {
	u.session = id
}

func (u *SessionBase) GetClientId() string {
	return u.client
}

func (u *SessionBase) SetClientId(id string) {
	u.client = id
}

type AuthDataAccessor interface {
	Set(key string, value string)
	Get(key string) string
}

type AuthContext interface {
	op_context.Context
	Session

	GetRequestContent() []byte
	CheckRequestContent(smsMessage *string) error
	GetRequestPath() string
	GetRequestMethod() string

	GetRequestClientIp() string
	GetRequestUserAgent() string

	AuthUser() User
	SetAuthUser(user User)

	SetAuthParameter(authMethodProtocol string, key string, value string)
	GetAuthParameter(authMethodProtocol string, key string) string

	Tenancy() multitenancy.Tenancy
}

type AuthHandler interface {
	common.WithName

	Protocol() string

	Handle(ctx AuthContext) (bool, error)
	Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error

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
