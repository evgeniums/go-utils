package api_server

// Interface of generic server that implements some API.
type Server interface {

	// Get API version.
	ApiVersion() string

	// Run server.
	Run()

	// Add a service to server.
	AddService(service Service)

	// Check if server supports multiple tenancies
	IsMultiTenancy() bool

	// Add tenancy.
	AddTenancy(tenancy Tenancy) error

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Remove tenance.
	RemoveTenancy(id string) error
}
