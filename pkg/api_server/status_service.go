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

type StatusService struct {
	ServiceBase
}

func NewStatusService() *StatusService {
	s := &StatusService{}
	s.GroupBase.Init("/status", "Status")
	s.AddEndpoints(NewCheckStatusEndpoint())
	return s
}
