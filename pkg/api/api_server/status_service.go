package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
)

type CheckStatusEndpoint struct {
	ResourceEndpoint
}

func NewCheckStatusEndpoint() *CheckStatusEndpoint {
	ep := &CheckStatusEndpoint{}
	ep.Init("check", "CheckStatus", ep, access_control.Get)
	return ep
}

type StatusResponse struct {
	Status string `json:"status"`
}

func (e *CheckStatusEndpoint) HandleRequest(request Request) error {
	resp := &StatusResponse{"running"}
	request.Response().SetMessage(resp)
	return nil
}

type CheckAccess struct{}

func (e *CheckAccess) HandleRequest(request Request) error {
	resp := &StatusResponse{"success"}
	request.Response().SetMessage(resp)
	return nil
}

type CheckAccessEndpoint struct {
	EndpointBase
	CheckAccess
}

func NewCheckAccessEndpoint(operationName string, accessType ...access_control.AccessType) *CheckAccessEndpoint {
	ep := &CheckAccessEndpoint{}
	ep.Init(operationName, accessType...)
	return ep
}

type CheckAccessResourceEndpoint struct {
	ResourceEndpoint
	CheckAccess
}

func NewCheckAccessResourceEndpoint(resource string, operationName string,
	accessType ...access_control.AccessType) *CheckAccessResourceEndpoint {
	ep := &CheckAccessResourceEndpoint{}
	ep.Init(resource, operationName, ep, accessType...)
	return ep
}

type StatusService struct {
	ServiceBase
}

func NewStatusService() *StatusService {
	s := &StatusService{}

	s.Init("status")
	s.AddChildren(NewCheckStatusEndpoint(),
		NewCheckAccessResourceEndpoint("csrf", "CheckCsrf"),
		NewCheckAccessResourceEndpoint("logged", "CheckLogged"),
	)
	altSmsPath := NewCheckAccessResourceEndpoint("sms-alt", "CheckSmsAlt", access_control.Post)
	altSmsPath.SetTestOnly(true)
	s.AddChild(altSmsPath)

	sms := api.NewResource("sms")
	sms.AddOperations(
		NewCheckAccessEndpoint("CheckSms", access_control.Post),
	)
	altSmsMethod := NewCheckAccessEndpoint("CheckSmsPut", access_control.Put)
	altSmsMethod.SetTestOnly(true)
	sms.AddOperation(altSmsMethod)
	s.AddChild(sms)
	return s
}
