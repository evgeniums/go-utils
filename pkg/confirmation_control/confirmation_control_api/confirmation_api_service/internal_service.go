package confirmation_api_service

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
)

type InternalEndpoint struct {
	service *ConfirmationInternalService
	api_server.EndpointBase
}

func (e *InternalEndpoint) Construct(service *ConfirmationInternalService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type ConfirmationInternalService struct {
	api_server.ServiceBase
	OperationResource api.Resource
	BaseUrl           string
	TokenTtl          int
}

func NewConfirmationInternalService(baseUrl string, tokenTtl int) *ConfirmationInternalService {

	s := &ConfirmationInternalService{BaseUrl: baseUrl, TokenTtl: tokenTtl}

	s.Init(confirmation_control_api.ServiceName, true)
	s.OperationResource = api.NewResource(confirmation_control_api.OperationResource)
	s.AddChild(s.OperationResource)
	s.OperationResource.AddOperation(PrepareOperation(s))

	return s
}

type PrepareOperationEndpoint struct {
	InternalEndpoint
}

func (e *PrepareOperationEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("ConfirmationInternalService.PrepareOperation")
	defer request.TraceOutMethod()

	// parse command
	cmd := &confirmation_control_api.PrepareOperationCmd{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return err
	}

	request.SetLoggerField("confirmation_id", cmd.Id)

	// save data in cache
	cacheToken := &confirmation_control_api.OperationCacheToken{Id: cmd.Id, Recipient: cmd.Recipient, FailedUrl: cmd.FailedUrl, Parameters: cmd.Parameters}
	cacheKey := confirmation_control_api.OperationIdCacheKey(cmd.Id)
	err = request.Cache().Set(cacheKey, cacheToken, e.service.TokenTtl)
	if err != nil {
		c.SetMessage("failed to store operation in cache")
		return err
	}

	// set response
	resp := &confirmation_control_api.PrepareOperationResponse{}
	tenancyQuery := ""
	if request.GetTenancy() != nil {
		tenancyQuery = fmt.Sprintf("&tenancy=%s", multitenancy.ContextTenancyPath(request))
	}
	resp.Url = fmt.Sprintf("%s?operation=%s%s", e.service.BaseUrl, cmd.Id, tenancyQuery)
	request.SetLoggerField("url", resp.Url)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func PrepareOperation(s *ConfirmationInternalService) *PrepareOperationEndpoint {
	e := &PrepareOperationEndpoint{}
	e.Construct(s, confirmation_control_api.PrepareOperation())
	return e
}
