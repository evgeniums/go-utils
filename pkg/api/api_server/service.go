package api_server

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

type ServiceEachEndpointHandler = func(ep Endpoint)

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	api.Resource

	Server() Server
	AttachToServer(server Server, multitenancy ...bool) error
}

type ServiceBase struct {
	api.ResourceBase
	generic_error.ErrorsExtenderBase
	server Server
}

func (s *ServiceBase) Init(pathName string) {
	s.ResourceBase.Init(pathName, api.ResourceConfig{Service: true})
}

func (s *ServiceBase) Server() Server {
	return s.server
}

func (s *ServiceBase) AttachToServer(server Server, multitenancy ...bool) error {
	s.server = server
	s.AttachToErrorManager(server)
	return s.EachOperation(func(op api.Operation) error {
		ep, ok := op.(Endpoint)
		if !ok {
			return fmt.Errorf("invalid opertaion type, must be endpoint: %s", op.Name())
		}
		server.AddEndpoint(ep, multitenancy...)
		return nil
	})
}
