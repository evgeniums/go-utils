package confirmation_control_api

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

const ServiceName string = "confirmation"

const OperationResource string = "operation"
const CallbackResource string = "callback"

func CheckConfirmation() api.Operation {
	return api.Post("check_confirmation")
}

func PrepareCheckConfirmation() api.Operation {
	return api.Get("prepare_check_confirmation")
}

func PrepareOperation() api.Operation {
	return api.Post("prepare_operation")
}

type PrepareOperationCmd struct {
	Id        string `json:"id" validate:"required,id" vmessage:"Operation ID must be specified"`
	Recipient string `json:"recipient" validate:"required" vmessage:"Recipient must be specified"`
	FailedUrl string `json:"failed_url" validate:"required,url" vmessage:"Invalid format of failed URL"`
}

type PrepareOperationResponse struct {
	api.ResponseStub
	Url string `json:"url"`
}

type CodeCmd struct {
	Code string `json:"code" validate:"required" vmessage:"Code must be specified"`
}

type PrepareCheckConfirmationResponse struct {
	api.ResponseStub
	FailedUrl string `json:"failed_url"`
}

type OperationCacheToken struct {
	Id        string `json:"id"`
	Recipient string `json:"recipient"`
	FailedUrl string `json:"failed_url"`
}

const ConfirmationCacheKey = "confirmation_service"

func OperationIdCacheKey(operationId string) string {
	return fmt.Sprintf("%s/%s", ConfirmationCacheKey, operationId)
}

func GetTokenFromCache(ctx auth.AuthContext) (*OperationCacheToken, error) {

	// setup
	c := ctx.TraceInMethod("GetTokenFromCache")
	defer ctx.TraceOutMethod()

	// get token from cache
	operationId := ctx.GetResourceId(OperationResource)
	ctx.SetLoggerField("cache_operation_id", operationId)
	cacheToken := &OperationCacheToken{}
	cacheKey := OperationIdCacheKey(operationId)
	found, err := ctx.Cache().Get(cacheKey, cacheToken)
	if err != nil {
		c.SetMessage("failed to get cache token")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return nil, err
	}
	if !found {
		c.SetMessage("cache token not found")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeNotFound)
		return nil, err
	}

	// done
	return cacheToken, nil
}

func CallbackConfirmation() api.Operation {
	return api.Post("callback_confirmation")
}

type CallbackConfirmationCmd struct {
	Id           string `json:"operation_id" validate:"required,id" vmessage:"Invalid operation ID"`
	CodeOrStatus string `json:"code_status" validate:"required" vmessage:"Code or status must be specified"`
}

type CallbackConfirmationResponse struct {
	api.ResponseStub
	Url string `json:"url"`
}
