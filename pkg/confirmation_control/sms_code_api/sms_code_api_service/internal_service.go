package sms_code_api_service

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/sms_code_api"
)

type InternalEndpoint struct {
	service *SmsCodeInternalService
	api_server.EndpointBase
}

func (e *InternalEndpoint) Construct(service *SmsCodeInternalService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type SmsCodeInternalService struct {
	api_server.ServiceBase
	OperationResource api.Resource
	BaseUrl           string
	TokenTtl          int
}

func NewSmsCodeInternalService(baseUrl string, tokenTtl int) *SmsCodeInternalService {

	s := &SmsCodeInternalService{BaseUrl: baseUrl, TokenTtl: tokenTtl}

	s.Init(sms_code_api.ServiceName)
	s.OperationResource = api.NewResource(sms_code_api.OperationResource)
	s.AddChild(s.OperationResource)
	s.OperationResource.AddOperation(PrepareOperation(s))

	return s
}

type PrepareOperationEndpoint struct {
	InternalEndpoint
}

func (e *PrepareOperationEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("sms_code_api_service.PrepareOperation")
	defer request.TraceOutMethod()

	// parse command
	cmd := &sms_code_api.Operation{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return err
	}

	// save phone in cache
	cacheToken := &OperationCacheToken{Phone: cmd.Phone, FailedUrl: cmd.FailedUrl}
	cacheKey := OperationIdCacheKey(cmd.Id)
	err = request.Cache().Set(cacheKey, cacheToken, e.service.TokenTtl)
	if err != nil {
		c.SetMessage("failed to store operation in cache")
		return err
	}

	// set response
	resp := &sms_code_api.PrepareOperationResponse{}
	resp.Url = fmt.Sprintf("%s/tenancy/%s/%s/%s/%s", e.service.BaseUrl, request.GetTenancy().GetID(), sms_code_api.ServiceName, sms_code_api.OperationResource, cmd.Id)
	request.SetLoggerField("url", resp.Url)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func PrepareOperation(s *SmsCodeInternalService) *PrepareOperationEndpoint {
	e := &PrepareOperationEndpoint{}
	e.Construct(s, sms_code_api.PrepareOperation())
	return e
}
