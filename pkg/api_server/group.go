package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

// Interface of group of API endpoints.
type Group interface {
	common.WithNameAndPath
	generic_error.ErrorsExtender

	// Add endpoints to group.
	AddEndpoints(ep ...Endpoint)

	// Get endpoints
	// Endpoints() []Endpoint

	// TODO implement adding subgroup.
	// AddGroup(group Group) Group
}

// Base type of group of API endpoints.
type GroupBase struct {
	common.WithNameAndPathParentBase
	generic_error.ErrorsExtenderBase

	endpoints []Endpoint
}

func (e *GroupBase) Init(path string, name string) {
	e.WithNameAndPathParentBase.Init(path, name)
}

func (g *GroupBase) addEndpoint(endpoint Endpoint) {
	g.WithNameAndPathParentBase.AddChild(endpoint)
	g.endpoints = append(g.endpoints, endpoint)
}

func (g *GroupBase) AddEndpoints(endpoints ...Endpoint) {
	for _, ep := range endpoints {
		g.addEndpoint(ep)
	}
}

// func (g *GroupBase) Endpoints() []Endpoint {
// 	return g.endpoints
// }

// func (g *GroupBase) AddGroup(group Group) {
// 	g.WithNameAndPathParentBase.AddChild(group)
// 	for _, ep := range group.Endpoints() {
// 		g.addEndpoint(ep)
// 	}
// }
