package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/utils"

// Interface of API endpoint.
type Endpoint interface {
	WithNameAndPath

	// Handle reqeust to server API.
	HandleRequest(request Request) error

	// Check if 2-factor authorization is enabled by default for this endpoint.
	Is2FaDefault() bool

	// Precheck request before asking for 2-factor authprization.
	PrecheckRequestBefore2Fa(request Request) error
}

// Base type for API endpoints.
type EndpointBase struct {
	WithNameAndPathBase

	enable2FaDefault bool
}

func (e *EndpointBase) Is2FaDefault() bool {
	return e.enable2FaDefault
}

func (e *EndpointBase) Init(path string, name string, enable2FaDefault ...bool) {
	e.WithNameAndPathBase.Init(name, path)
	e.enable2FaDefault = utils.OptionalArg(false, enable2FaDefault...)
}
