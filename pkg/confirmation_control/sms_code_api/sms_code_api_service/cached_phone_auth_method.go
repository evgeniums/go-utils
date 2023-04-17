package sms_code_api_service

import (
	"fmt"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/sms_code_api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const ErrorCodeOperationNotFound = "operation_not_found"
const OperationCacheKey = "sms_code_service"

func OperationIdCacheKey(operationId string) string {
	return fmt.Sprintf("%s/%s", OperationCacheKey, operationId)
}

type OperationCacheToken struct {
	Phone     string `json:"phone"`
	FailedUrl string `json:"failed_url"`
}

type User struct {
	auth.UserBase
	phone string
}

func (u *User) Phone() string {
	return u.phone
}

const CachedPhoneAuthProtocol = "cached_phone"

type CachedPhoneAuthMethod struct {
	auth.AuthHandlerBase
}

func NewCachedPhoneAuthMethod() *CachedPhoneAuthMethod {
	c := &CachedPhoneAuthMethod{}
	return c
}

func (a *CachedPhoneAuthMethod) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	a.AuthHandlerBase.Init(CachedPhoneAuthProtocol)

	return nil
}

func (a *CachedPhoneAuthMethod) Handle(ctx auth.AuthContext) (bool, error) {

	c := ctx.TraceInMethod("CachedPhoneAuthMethod.Handle")
	defer ctx.TraceOutMethod()

	operationId := ctx.GetResourceId(sms_code_api.OperationResource)
	ctx.SetLoggerField("cache_operation_id", operationId)
	cacheToken := &OperationCacheToken{}
	cacheKey := OperationIdCacheKey(operationId)
	found, err := ctx.Cache().Get(cacheKey, cacheToken)
	if err != nil {
		c.SetMessage("failed to get cache token")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}
	if !found {
		c.SetMessage("cache token not found")
		ctx.SetGenericErrorCode(ErrorCodeOperationNotFound)
		return true, err
	}

	user := &User{}
	user.UserId = operationId
	user.UserDisplay = operationId
	user.UserLogin = "_none_"
	user.phone = cacheToken.Phone
	ctx.SetAuthUser(user)

	ctx.Cache().Unset(cacheKey)

	return true, nil
}

func (a *CachedPhoneAuthMethod) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(a)
}

func (a *CachedPhoneAuthMethod) ErrorDescriptions() map[string]string {
	m := map[string]string{
		ErrorCodeOperationNotFound: "Operation with such ID not found or expired.",
	}
	return m
}

func (a *CachedPhoneAuthMethod) ErrorProtocolCodes() map[string]int {
	m := map[string]int{
		ErrorCodeOperationNotFound: http.StatusNotFound,
	}
	return m
}
