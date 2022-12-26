package api_server

// Interface of group of API endpoints.
type Group interface {
	WithNameAndPath

	// Add endpoint to the group.
	AddEndpoint(endpoint Endpoint) error

	// Implement this method in derived types.
	DoAddEndpoint(endpoint Endpoint) error
}

// Base type of group of API endpoints.
type GroupBase struct {
	Group
	WithNameAndPathBase
}

// Add endpoint to the group.
func (g *GroupBase) AddEndpoint(endpoint Endpoint) error {
	endpoint.setParentPath(g.path)
	return g.DoAddEndpoint(endpoint)
}
