package auth_factory

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_hmac"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_token"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/noauth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
)

type DefaultAuthFactory struct {
	Users      auth_session.WithUserSessionManager
	SmsManager sms.SmsManager
}

func (f *DefaultAuthFactory) Create(protocol string) (auth.AuthHandler, error) {

	switch protocol {
	case LoginphashTokenProtocol:
		return NewLoginphashToken(f.Users), nil
	case LoginphashSmsTokenProtocol:
		return NewLoginphashSmsToken(f.Users, f.SmsManager), nil
	case auth_login_phash.LoginProtocol:
		return auth_login_phash.New(f.Users), nil
	case auth_token.CheckTokenProtocol:
		return auth_token.New(f.Users), nil
	case auth_token.TokenProtocol:
		return auth_token.NewSchema(f.Users), nil
	case auth_hmac.HmacProtocol:
		return &auth_hmac.AuthHmac{}, nil
	case auth_sms.SmsProtocol:
		return auth_sms.New(f.SmsManager), nil
	case noauth.NoAuthProtocol:
		return &noauth.NoAuth{}, nil

	}

	return nil, fmt.Errorf("unknown auth handler %s", protocol)
}
