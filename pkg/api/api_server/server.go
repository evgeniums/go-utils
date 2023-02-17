package api_server

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	finish "github.com/evgeniums/go-finish-service"
)

// Interface of generic server that implements some API.
type Server interface {
	generic_error.ErrorManager
	auth.WithAuth

	// Get API version.
	ApiVersion() string

	// Run server.
	Run(fin *finish.Finisher)

	// Add operation endpoint to server.
	AddEndpoint(ep Endpoint)

	// Check if hateoas links are enabled.
	IsHateoas() bool

	// Get tenancy manager
	TenancyManager() multitenancy.Multitenancy

	// Check for testing mode.
	Testing() bool
}

func AddServiceToServer(s Server, service Service) {
	err := service.AttachToServer(s)
	if err != nil {
		panic(fmt.Errorf("failed to attach service %s to server", service.Type()))
	}
}

type ServerBaseConfig struct {
	common.WithNameBaseConfig
	API_VERSION     string `validate:"required"`
	HATEOAS         bool
	OPLOG_USER_TYPE string
}

func (s *ServerBaseConfig) ApiVersion() string {
	return s.API_VERSION
}

func (s *ServerBaseConfig) IsHateoas() bool {
	return s.HATEOAS
}
