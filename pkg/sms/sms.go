package sms

import (
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ProviderResponse struct {
	ProviderMessageID string
	RawContent        string
}

type Provider interface {
	Protocol() string
	Name() string
	Send(ctx op_context.Context, message string, recipient string, smsID ...string) (*ProviderResponse, error)
}

type ProviderBase struct {
	protocol string
	name     string
}

func (p *ProviderBase) Init(protocol string, name ...string) {
	p.protocol = protocol
	p.name = utils.OptionalArg(protocol, name...)
}

func (p *ProviderBase) Name() string {
	return p.name
}

func (p *ProviderBase) Protocol() string {
	return p.protocol
}
