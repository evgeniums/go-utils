package api_server

// Interface of generic server that implements some API.
type Server interface {

	// Get API version.
	ApiVersion() string

	// Run server.
	Run()

	// Add a service to server.
	AddService(service Service)
}
