package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	finish "github.com/evgeniums/go-finish-service"
)

// Interface of generic server that implements some API.
type Server interface {

	// Get API version.
	ApiVersion() string

	// Run server.
	Run(fin *finish.Finisher)

	// Add endpoint to server.
	AddEndpoint(ep Endpoint)

	// Check if server supports multiple tenancies
	IsMultiTenancy() bool

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Add tenancy.
	AddTenancy(id string) error

	// Remove tenance.
	RemoveTenancy(id string) error
}

func AddServiceToServer(s Server, service Service) {
	service.AttachToServer(s)
}

type ServerBaseConfig struct {
	API_VERSION  string `validate:"required"`
	MULTITENANCY bool
	common.WithNameBaseConfig
}

func (s *ServerBaseConfig) ApiVersion() string {
	return s.API_VERSION
}

func (s *ServerBaseConfig) IsMultiTenancy() bool {
	return s.MULTITENANCY
}
