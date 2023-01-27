package sms_provider_factory

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/providers/gatewayapi"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/providers/smsru"
)

type DefaultFactory struct{}

func (f *DefaultFactory) Create(protocol string) (sms.Provider, error) {

	switch protocol {
	case gatewayapi.Protocol:
		return gatewayapi.New(), nil
	case smsru.Protocol:
		return smsru.New(), nil
	}

	return nil, errors.New("unknown SMS provider")
}
