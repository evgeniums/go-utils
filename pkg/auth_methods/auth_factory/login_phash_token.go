package auth_factory

import (
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_token"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const LoginphashTokenProtocol = "login_phash_token"

type LoginphashToken struct {
	auth.AuthSchema

	Login *auth_login_phash.LoginHandler
	Token *auth_token.AuthNewTokenHandler
}

func NewLoginphashToken(users user_manager.WithSessionManager) *LoginphashToken {
	l := &LoginphashToken{}
	l.Construct()
	l.Login = auth_login_phash.New(users)
	l.Token = auth_token.NewNewToken(users)
	return l
}

func (l *LoginphashToken) InitLoginToken(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {
	l.AuthSchema.SetAggregation(auth.And)

	path := utils.OptionalArg("auth_manager.methods", configPath...)

	pathParts := strings.Split(path, ".")
	pathParts = pathParts[:len(pathParts)-1]

	parentPath := strings.Join(pathParts, ".")

	loginCfgPath := object_config.Key(parentPath, auth_login_phash.LoginProtocol)
	tokenCfgPath := object_config.Key(parentPath, auth_token.TokenProtocol)

	err := l.Login.Init(cfg, log, vld, loginCfgPath)
	if err != nil {
		return log.PushFatalStack("failed to init login handler", err)
	}

	err = l.Token.Init(cfg, log, vld, tokenCfgPath)
	if err != nil {
		return log.PushFatalStack("failed to init token handler", err)
	}

	return nil
}

func (l *LoginphashToken) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	l.AuthHandlerBase.Init(LoginphashTokenProtocol)

	err := l.InitLoginToken(cfg, log, vld, configPath...)
	if err != nil {
		return err
	}

	l.AuthSchema.AppendHandlers(l.Login, l.Token)
	return nil
}

func (l *LoginphashToken) Handlers() []auth.AuthHandler {
	return l.AuthSchema.Handlers()
}

func (l *LoginphashToken) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(l)
}
