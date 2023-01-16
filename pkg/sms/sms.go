package sms

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Provider interface {
	Send(ctx op_context.Context, message string, recipient string, smsID ...string) (string, error)
}

type SmsManager interface {
	generic_error.ErrorDefinitions

	Send(ctx op_context.Context, message string, recipient string) (string, error)
}

const (
	ErrorCodeSmsSendingFailed string = "sms_sending_failed"
)

var SmsErrorDescriptions = map[string]string{
	ErrorCodeSmsSendingFailed: "Failed to send SMS",
}

var SmsErrorHttpCodes = map[string]int{
	ErrorCodeSmsSendingFailed: http.StatusInternalServerError,
}
