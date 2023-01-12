package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	finish "github.com/evgeniums/go-finish-service"
)

// Interface of generic server that implements some API.
type Server interface {
	generic_error.ErrorManager
	multitenancy.Multitenancy

	// Get API version.
	ApiVersion() string

	// Run server.
	Run(fin *finish.Finisher)

	// Add endpoint to server.
	AddEndpoint(ep Endpoint)
}

func AddServiceToServer(s Server, service Service) {
	service.AttachToServer(s)
}

type ServerBaseConfig struct {
	API_VERSION string `validate:"required"`
	common.WithNameBaseConfig
	multitenancy.MultitenancyBaseConfig
}

func (s *ServerBaseConfig) ApiVersion() string {
	return s.API_VERSION
}
