package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/generic_error"

type ServiceEachEndpointHandler = func(ep Endpoint)

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	generic_error.ErrorsExtender
	Group

	// Attach service to server.
	AttachToServer(server Server)
}

type ServiceBase struct {
	generic_error.ErrorsExtenderBase
	GroupBase
}

func (s *ServiceBase) AttachToServer(server Server) {
	s.AddToErrorManager(server)
	for _, ep := range s.endpoints {
		server.AddEndpoint(ep)
	}
}
