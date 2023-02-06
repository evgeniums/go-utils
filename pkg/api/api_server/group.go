package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

// Interface of group of API endpoints.
type Group interface {
	common.WithNameAndPath
	generic_error.ErrorsExtender

	// Add endpoints to group.
	AddEndpoints(ep ...Endpoint)

	// Get all endpoints including endpoints in subgroups.
	Endpoints(direct ...bool) []Endpoint

	Groups() []Group

	Parent() Group

	// Add subgroup.
	AddGroup(group Group)

	clearEndpoints()
	clearGroups()
}

type WithParentBase struct {
}

// Base type of group of API endpoints.
type GroupBase struct {
	common.WithNameAndPathParentBase
	generic_error.ErrorsExtenderBase

	endpoints map[string]Endpoint
	groups    map[string]Group

	parent Group
}

func (e *GroupBase) Init(path string, name string) {
	e.WithNameAndPathParentBase.Init(path, name)
	e.endpoints = make(map[string]Endpoint)
	e.groups = make(map[string]Group)
}

func (g *GroupBase) addEndpoint(ep Endpoint) {
	g.AddChild(ep)
	g.endpoints[ep.Path()] = ep
}

func (g *GroupBase) AddEndpoints(endpoints ...Endpoint) {
	for _, ep := range endpoints {
		g.addEndpoint(ep)
	}
}

func (g *GroupBase) Endpoints(direct ...bool) []Endpoint {
	eps := make([]Endpoint, len(g.endpoints))
	i := 0
	for _, ep := range g.endpoints {
		eps[i] = ep
		i++
	}

	directOnly := utils.OptionalArg(false, direct...)
	if !directOnly {
		for _, group := range g.groups {
			eps = append(eps, group.Endpoints()...)
		}
	}
	return eps
}

func (g *GroupBase) AddGroup(group Group) {

	// add group
	g.AddChild(group)
	g.groups[group.Path()] = group

	// re-add endpoints
	eps := group.Endpoints(true)
	group.clearEndpoints()
	group.AddEndpoints(eps...)

	// re-add groups
	groups := group.Groups()
	group.clearGroups()
	for _, gr := range groups {
		group.AddGroup(gr)
	}
}

func (g *GroupBase) clearEndpoints() {
	g.endpoints = make(map[string]Endpoint)
}

func (g *GroupBase) clearGroups() {
	g.groups = make(map[string]Group)
}

func (g *GroupBase) Groups() []Group {
	grs := make([]Group, len(g.groups))
	i := 0
	for _, gr := range g.groups {
		grs[i] = gr
		i++
	}
	return grs
}

func (g *GroupBase) Parent() Group {
	return g.parent
}

func (g *GroupBase) SetParent(parent common.WithPath) {
	g.parent = parent.(Group)
	g.WithPathBase.SetParent(parent)
}
