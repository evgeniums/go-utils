package api_server

// Interface of service that implements one or more groups of API endpoints.
type Service interface {
	WithNameAndPath

	// Add endpoint to default service group.
	AddEndpoint(endpoint Endpoint) error
}
