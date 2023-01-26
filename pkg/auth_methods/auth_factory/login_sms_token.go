package auth_factory

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_token"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const LoginphashSmsTokenProtocol = "login_phash_sms_token"

type LoginphashSmsToken struct {
	LoginphashToken
	Sms auth.AuthHandler
}

func NewLoginphashSmsToken(users user_manager.WithSessionManager) *LoginphashSmsToken {
	l := &LoginphashSmsToken{}
	l.Login = auth_login_phash.New(users)
	l.Token = auth_token.New(users)
	return l
}

func (l *LoginphashSmsToken) InitSms(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	path := utils.OptionalArg("auth_manager.methods", configPath...)
	smsCfgPath := object_config.Key(path, auth_sms.SmsProtocol)

	sms := &auth_sms.AuthSms{}
	err := sms.Init(cfg, log, vld, smsCfgPath)
	if err != nil {
		return log.PushFatalStack("failed to init SMS handler", err)
	}
	l.Sms = sms

	return nil
}

func (l *LoginphashSmsToken) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	l.AuthHandlerBase.Init(LoginphashSmsTokenProtocol)

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
