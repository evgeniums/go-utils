package api_server

type ServiceEachEndpointHandler = func(ep Endpoint)

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	Group

	// Attach service to server.
	AttachToServer(server Server)
}

type ServiceBase struct {
	GroupBase
}

func (s *ServiceBase) AttachToServer(server Server) {
	s.AttachToErrorManager(server)
	for _, ep := range s.endpoints {
		server.AddEndpoint(ep)
	}
}
