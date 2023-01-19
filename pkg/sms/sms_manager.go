package sms

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

type SmsManager interface {
	generic_error.ErrorDefinitions

	Send(ctx auth.AuthContext, message string, recipient string) (string, error)
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
