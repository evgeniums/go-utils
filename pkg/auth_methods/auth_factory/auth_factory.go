package auth_factory

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_hmac"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_sms"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_token"
)

type DefaultAuthFactory struct {
}

func (f *DefaultAuthFactory) Create(protocol string) (auth.AuthHandler, error) {

	switch protocol {
	case LoginphashTokenProtocol:
		return &LoginphashToken{}, nil
	case LoginphashSmsTokenProtocol:
		return &LoginphashSmsToken{}, nil
	case auth_login_phash.LoginProtocol:
		return &auth_login_phash.LoginHandler{}, nil
	case auth_token.TokenProtocol:
		return &auth_token.AuthTokenHandler{}, nil
	case auth_hmac.HmacProtocol:
		return &auth_hmac.AuthHmac{}, nil
	case auth_sms.SmsProtocol:
		return &auth_hmac.AuthHmac{}, nil

	}

	return nil, fmt.Errorf("unknown auth handler %s", protocol)
}
