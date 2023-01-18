package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
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

	// Precheck request before some authorization methods
	PrecheckRequestBeforeAuth(request Request, authDataAccessor ...auth.AuthDataAccessor) error
}

type EndpointHandler = func(request Request)

// Base type for API endpoints.
type EndpointBase struct {
	common.WithNameAndPathBase

	service    Service
	accessType access_control.AccessType
}

func (e *EndpointBase) Init(path string, name string, accessType access_control.AccessType) {
	e.WithNameAndPathBase.Init(name, path)
	e.accessType = accessType
}

func (e *EndpointBase) Service() Service {
	return e.service
}

func (e *EndpointBase) SetService(service Service) {
	e.service = service
}

func (e *EndpointBase) AccessType() access_control.AccessType {
	return e.accessType
}

func (e *EndpointBase) PrecheckRequestBeforeAuth(request Request, authDataAccessor ...auth.AuthDataAccessor) error {
	return nil
}

// Base type for API endpoints with empty handlers.
type EndpointNoHandler struct {
	EndpointBase
}

func (e *EndpointNoHandler) HandleRequest(request Request) error {
	return nil
}
