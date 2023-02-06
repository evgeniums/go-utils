package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

// Interface of API endpoint.
type Endpoint interface {
	common.WithNameAndPath
	generic_error.ErrorsExtender

	Parent() Group

	// Get service
	Service() Service
	// Set service
	SetService(service Service)

	// Get endpoint access type
	AccessType() access_control.AccessType

	// Handle request to server API.
	HandleRequest(request Request) error

	// Precheck request before some authorization methods
	PrecheckRequestBeforeAuth(request Request, smsMessage *string) error
}

type EndpointHandler = func(request Request)

// Base type for API endpoints.
type EndpointBase struct {
	common.WithNameAndPathBase
	generic_error.ErrorsExtenderBase

	service    Service
	accessType access_control.AccessType
	parent     Group
}

func (e *EndpointBase) Init(path string, name string, accessType access_control.AccessType) {
	e.WithNameAndPathBase.Init(path, name)
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

func (e *EndpointBase) PrecheckRequestBeforeAuth(request Request, smsMessage *string) error {
	return nil
}

func (e *EndpointBase) Parent() Group {
	return e.parent
}

func (e *EndpointBase) SetParent(parent common.WithPath) {
	// TODO fix parent
	// e.parent = parent.(Group)
	e.WithNameAndPathBase.SetParent(parent)
}

// Base type for API endpoints with empty handlers.
type EndpointNoHandler struct {
	EndpointBase
}

func (e *EndpointNoHandler) HandleRequest(request Request) error {
	return nil
}
