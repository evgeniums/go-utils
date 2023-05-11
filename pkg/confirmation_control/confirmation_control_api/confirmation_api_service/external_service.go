package confirmation_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api"
)

type ExternalEndpoint struct {
	service *ConfirmationExternalService
	api_server.EndpointBase
}

func (e *ExternalEndpoint) Construct(service *ConfirmationExternalService, op api.Operation) {
	e.service = service
	e.EndpointBase.Construct(op)
}

type ConfirmationExternalService struct {
	api_server.ServiceBase
	ConfirmationCallbackHandler confirmation_control.ConfirmationCallbackHandler

	OperationResource api.Resource

	CheckCode bool
}

func NewConfirmationExternalService(confirmationCallbackHandler confirmation_control.ConfirmationCallbackHandler, checkCode bool) *ConfirmationExternalService {

	s := &ConfirmationExternalService{CheckCode: checkCode}
	s.ConfirmationCallbackHandler = confirmationCallbackHandler

	s.Init(confirmation_control_api.ServiceName, true)
	s.OperationResource = api.NamedResource(confirmation_control_api.OperationResource)
	s.AddChild(s.OperationResource.Parent())
	s.OperationResource.AddOperations(CheckConfirmation(s), PrepareCheckConfirmation(s))

	return s
}

type CheckConfirmationEndpoint struct {
	ExternalEndpoint
}

func (e *CheckConfirmationEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	var err error
	c := request.TraceInMethod("ConfirmationExternalService.CheckConfirmation")
	defer request.TraceOutMethod()

	// fill code or status
	var codeOrStatus string
	if e.service.CheckCode {
		// parse command
		cmd := &confirmation_control_api.CheckConfirmationCmd{}
		err := request.ParseValidate(cmd)
		if err != nil {
			c.SetMessage("failed to parse/validate command")
			return err
		}
		codeOrStatus = cmd.Code
	} else {
		codeOrStatus = confirmation_control.StatusSuccess
	}

	// invoke callback
	resp := &confirmation_control_api.CheckConfirmationResponse{}
	resp.RedirectUrl, err = e.service.ConfirmationCallbackHandler.ConfirmationCallback(request, request.GetResourceId(confirmation_control_api.OperationResource), codeOrStatus)
	if err != nil {
		c.SetMessage("failed to invoke callback")
		return c.SetError(err)
	}

	// set response
	request.Response().SetMessage(resp)

	// done
	return nil
}

func CheckConfirmation(s *ConfirmationExternalService) *CheckConfirmationEndpoint {
	e := &CheckConfirmationEndpoint{}
	e.Construct(s, confirmation_control_api.CheckConfirmation())
	return e
}

type PrepareCheckConfirmationEndpoint struct {
	ExternalEndpoint
}

func (e *PrepareCheckConfirmationEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("ConfirmationExternalService.PrepareCheckConfirmation")
	defer request.TraceOutMethod()

	// get token from cache
	cacheToken, err := confirmation_control_api.GetTokenFromCache(request)
	if err != nil {
		return c.SetError(err)
	}

	// set response
	resp := &confirmation_control_api.PrepareCheckConfirmationResponse{}
	resp.FailedUrl = cacheToken.FailedUrl
	resp.CodeInBody = e.service.CheckCode
	request.SetLoggerField("failed_url", resp.FailedUrl)
	request.Response().SetMessage(resp)

	// done
	return nil
}

func PrepareCheckConfirmation(s *ConfirmationExternalService) *PrepareCheckConfirmationEndpoint {
	e := &PrepareCheckConfirmationEndpoint{}
	e.Construct(s, confirmation_control_api.PrepareCheckConfirmation())
	return e
}
