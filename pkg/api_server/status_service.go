package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type CheckStatusEndpoint struct {
	EndpointBase
}

func NewCheckStatusEndpoint() *CheckStatusEndpoint {
	ep := &CheckStatusEndpoint{}
	ep.EndpointBase.Init("/check", "CheckStatus", access_control.Get)
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

type CheckAccessEndpoint struct {
	EndpointBase
}

func NewCheckAccessEndpoint(path string, name string, access ...access_control.AccessType) *CheckAccessEndpoint {
	accessType := utils.OptionalArg(access_control.Get, access...)
	ep := &CheckAccessEndpoint{}
	ep.EndpointBase.Init(path, name, accessType)
	return ep
}

func (e *CheckAccessEndpoint) HandleRequest(request Request) error {
	resp := &StatusResponse{"success"}
	request.Response().SetMessage(resp)
	return nil
}

type StatusService struct {
	ServiceBase
}

// TODO Endpoint at the same path but different access methods
func NewStatusService() *StatusService {
	s := &StatusService{}
	s.GroupBase.Init("/status", "Status")
	s.AddEndpoints(NewCheckStatusEndpoint(),
		NewCheckAccessEndpoint("/csrf", "CheckCsrf"),
		NewCheckAccessEndpoint("/logged", "CheckLogged"),
		NewCheckAccessEndpoint("/sms", "CheckSms", access_control.Post),
		// NewCheckAccessEndpoint("/sms", "CheckSmsPut", access_control.Put),
		NewCheckAccessEndpoint("/sms-alt", "CheckSmsAlt", access_control.Post),
	)
	return s
}
