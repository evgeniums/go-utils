package sms_code_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/sms_code_api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

type ExternalEndpoint struct {
	service *SmsCodeExternalService
	api_server.EndpointBase
}

func (e *ExternalEndpoint) Construct(service *SmsCodeExternalService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type SmsCodeExternalService struct {
	api_server.ServiceBase
	ConfirmationCallbackHandler confirmation_control.ConfirmationCallbackHandler

	OperationResource api.Resource
}

func NewSmsCodeExternalService(confirmationCallbackHandler confirmation_control.ConfirmationCallbackHandler) *SmsCodeExternalService {

	s := &SmsCodeExternalService{}
	s.ConfirmationCallbackHandler = confirmationCallbackHandler

	s.Init(sms_code_api.ServiceName)
	s.OperationResource = api.NamedResource(sms_code_api.OperationResource)
	s.AddChild(s.OperationResource.Parent())
	s.OperationResource.AddOperations(CheckSms(s), PrepareCheckSms(s))

	return s
}

type CheckSmsEndpoint struct {
	ExternalEndpoint
}

func (e *CheckSmsEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	var err error
	c := request.TraceInMethod("sms_code_api_service.CheckSms")
	defer request.TraceOutMethod()

	// invoke callback
	redirectUrl, err := e.service.ConfirmationCallbackHandler.ConfirmationCallback(request, request.GetResourceId(sms_code_api.OperationResource), confirmation_control.StatusSuccess)
	if redirectUrl != "" {
		request.Response().SetRedirectPath(redirectUrl)
	}
	if err != nil {
		c.SetMessage("failed to invoke callback")
		return c.SetError(err)
	}

	// done
	return nil
}

func CheckSms(s *SmsCodeExternalService) *CheckSmsEndpoint {
	e := &CheckSmsEndpoint{}
	e.Construct(s, sms_code_api.CheckSms())
	return e
}

type PrepareCheckSmsEndpoint struct {
	ExternalEndpoint
}

func (e *PrepareCheckSmsEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	var err error
	c := request.TraceInMethod("sms_code_api_service.PrepareCheckSms")
	defer request.TraceOutMethod()

	operationId := request.GetResourceId(sms_code_api.OperationResource)
	request.SetLoggerField("cache_operation_id", operationId)
	cacheToken := &OperationCacheToken{}
	cacheKey := OperationIdCacheKey(operationId)
	found, err := request.Cache().Get(cacheKey, cacheToken)
	if err != nil {
		c.SetMessage("failed to get cache token")
		request.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}
	if !found {
		c.SetMessage("cache token not found")
		request.SetGenericErrorCode(ErrorCodeOperationNotFound)
		return err
	}

	// set response
	resp := &sms_code_api.PrepareCheckSmsResponse{}
	resp.FailedUrl = cacheToken.FailedUrl
	request.SetLoggerField("url", resp.FailedUrl)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func PrepareCheckSms(s *SmsCodeExternalService) *PrepareCheckSmsEndpoint {
	e := &PrepareCheckSmsEndpoint{}
	e.Construct(s, sms_code_api.PrepareCheckSms())
	return e
}
