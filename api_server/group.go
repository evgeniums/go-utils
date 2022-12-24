package api_server

type Group interface {
	WithNameAndPath

	AddEndpoint(endpoint Endpoint) error
	DoAddEndpoint(endpoint Endpoint) error
}

type GroupBase struct {
	Group
	WithNameAndPathBase
}

func (g *GroupBase) AddEndpoint(endpoint Endpoint) error {
	endpoint.setParentPath(g.path)
	return g.DoAddEndpoint(endpoint)
}
