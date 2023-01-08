package api_server

import finish "github.com/evgeniums/go-finish-service"

// Interface of generic server that implements some API.
type Server interface {

	// Get API version.
	ApiVersion() string

	// Run server.
	Run(fin *finish.Finisher)

	// Add a service to server.
	AddService(service Service)

	// Check if server supports multiple tenancies
	IsMultiTenancy() bool

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Add tenancy.
	AddTenancy(id string) error

	// Remove tenance.
	RemoveTenancy(id string) error
}
