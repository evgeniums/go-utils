package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

// Interface of API endpoint.
type Endpoint interface {
	common.WithNameAndPath

	// Get service
	Service() Service
	// Set service
	SetService(service Service)

	// Get endpoint access type
	AccessType() access_control.AccessType

	// Handle request to server API.
	HandleRequest(request Request) error

	// Check if 2-factor authorization is enabled by default for this endpoint.
	Is2FaDefault() bool

	// Precheck request before asking for 2-factor authorization.
	PrecheckRequestBefore2Fa(request Request) error
}

type EndpointHandler = func(request Request)

// Base type for API endpoints.
type EndpointBase struct {
	common.WithNameAndPathBase

	enable2FaDefault bool
	service          Service
}

func (e *EndpointBase) Is2FaDefault() bool {
	return e.enable2FaDefault
}

func (e *EndpointBase) Init(path string, name string, enable2FaDefault ...bool) {
	e.WithNameAndPathBase.Init(name, path)
	e.enable2FaDefault = utils.OptionalArg(false, enable2FaDefault...)
}

func (e *EndpointBase) Service() Service {
	return e.service
}

func (e *EndpointBase) SetService(service Service) {
	e.service = service
}
