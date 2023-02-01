package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
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

func NewCheckAccessEndpoint(path string, name string) *CheckAccessEndpoint {
	ep := &CheckAccessEndpoint{}
	ep.EndpointBase.Init(path, name, access_control.Get)
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

func NewStatusService() *StatusService {
	s := &StatusService{}
	s.GroupBase.Init("/status", "Status")
	s.AddEndpoints(NewCheckStatusEndpoint(), NewCheckAccessEndpoint("/csrf", "CheckCsrf"), NewCheckAccessEndpoint("/logged", "CheckLogged"))
	return s
}
