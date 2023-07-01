package confirmation_control

import (
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
)

const StatusSuccess string = "success"
const StatusFailed string = "failed"
const StatusCancelled string = "cancelled"

type ConfirmationSender interface {
	SendConfirmation(ctx multitenancy.TenancyContext, operationId string, recipient string, failedUrl string, parameters ...map[string]interface{}) (redirectUrl string, err error)
}

type ConfirmationResult struct {
	Code   string                   `json:"code,omitempty"`
	Status string                   `json:"status,omitempty"`
	Error  *generic_error.ErrorBase `json:"error,omitempty"`
}

type ConfirmationCallbackHandler interface {
	ConfirmationCallback(ctx multitenancy.TenancyContext, operationId string, result *ConfirmationResult) (redirectUrl string, err error)
}
