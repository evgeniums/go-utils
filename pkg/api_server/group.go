package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/common"

// Interface of group of API endpoints.
type Group interface {
	common.WithNameAndPathParent

	// Add endpoint to the group.
	AddEndpoint(endpoint Endpoint) error

	// Implement this method in derived types.
	DoAddEndpoint(endpoint Endpoint) error
}

// Base type of group of API endpoints.
type GroupBase struct {
	Group
	common.WithNameAndPathParentBase
}

// Add endpoint to the group.
func (g *GroupBase) AddEndpoint(endpoint Endpoint) error {
	g.WithNameAndPathParentBase.AddChild(endpoint)
	return g.DoAddEndpoint(endpoint)
}
