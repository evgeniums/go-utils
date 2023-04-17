package confirmation_control

import "github.com/evgeniums/go-backend-helpers/pkg/multitenancy"

const StatusSuccess string = "success"

type ConfirmationSender interface {
	Send(ctx multitenancy.TenancyContext, operationId string, recipient string, failedUrl string) (redirectUrl string, err error)
}

type ConfirmationCallbackHandler interface {
	ConfirmationCallback(ctx multitenancy.TenancyContext, operationId string, codeOrStatus string, callbackUrl ...string) (redirectUrl string, err error)
}
