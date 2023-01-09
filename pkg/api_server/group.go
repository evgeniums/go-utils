package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/common"

// Interface of group of API endpoints.
type Group interface {
	common.WithNameAndPath

	// Add endpoints to group.
	AddEndpoint(ep ...Endpoint) Group

	// Add subgroup.
	AddGroup(group Group) Group
}

// Base type of group of API endpoints.
type GroupBase struct {
	common.WithNameAndPathParentBase
	endpoints []Endpoint
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
