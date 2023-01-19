package sms

import (
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type ProviderResponse struct {
	ProviderMessageID string
	RawContent        string
}

type Provider interface {
	Send(ctx op_context.Context, message string, recipient string, smsID ...string) (*ProviderResponse, error)
}
