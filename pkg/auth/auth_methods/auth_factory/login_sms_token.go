package auth_factory

import (
	"github.com/evgeniums/go-utils/pkg/auth"
	"github.com/evgeniums/go-utils/pkg/auth/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-utils/pkg/auth/auth_methods/auth_sms"
	"github.com/evgeniums/go-utils/pkg/auth/auth_methods/auth_token"
	"github.com/evgeniums/go-utils/pkg/auth/auth_session"
	"github.com/evgeniums/go-utils/pkg/config"
	"github.com/evgeniums/go-utils/pkg/config/object_config"
	"github.com/evgeniums/go-utils/pkg/logger"
	"github.com/evgeniums/go-utils/pkg/sms"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/evgeniums/go-utils/pkg/validator"
)

const LoginphashSmsTokenProtocol = "login_phash_sms_token"

type LoginphashSmsToken struct {
	LoginphashToken
	Sms auth.AuthHandler
}

func NewLoginphashSmsToken(users auth_session.WithUserSessionManager, smsManager sms.SmsManager) *LoginphashSmsToken {
	l := &LoginphashSmsToken{}
	l.Login = auth_login_phash.New(users)
	l.Token = auth_token.NewNewToken(users)
	l.Sms = auth_sms.New(smsManager)
	return l
}

func (l *LoginphashSmsToken) InitSms(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	path := utils.OptionalArg("auth_manager.methods", configPath...)
	smsCfgPath := object_config.Key(path, auth_sms.SmsProtocol)

	err := l.Sms.Init(cfg, log, vld, smsCfgPath)
	if err != nil {
		return log.PushFatalStack("failed to init SMS handler", err)
	}

	return nil
}

func (l *LoginphashSmsToken) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	l.AuthHandlerBase.Init(LoginphashSmsTokenProtocol)
	l.SetName(LoginphashTokenProtocol)

	err := l.InitLoginToken(cfg, log, vld, configPath...)
	if err != nil {
		return err
	}

	err = l.InitSms(cfg, log, vld, configPath...)
	if err != nil {
		return err
	}

	l.AuthSchema.AppendHandlers(l.Login, l.Sms, l.Token)
	return nil
}

func (l *LoginphashSmsToken) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(l)
}
