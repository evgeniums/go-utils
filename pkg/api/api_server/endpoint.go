package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

// Interface of API endpoint.
type Endpoint interface {
	api.Operation
	generic_error.ErrorsExtender

	// Handle request to server API.
	HandleRequest(request Request) error

	// Precheck request before some authorization methods
	PrecheckRequestBeforeAuth(request Request, smsMessage *string) error
}

type EndpointHandler = func(request Request)

// Base type for API endpoints.
type EndpointBase struct {
	api.Operation
	generic_error.ErrorsExtenderBase
}

func (e *EndpointBase) Construct(op api.Operation) {
	e.Operation = op
}

func (e *EndpointBase) Init(operationName string, accessType ...access_control.AccessType) {
	e.Construct(api.NewOperation(operationName, utils.OptionalArg(access_control.Get, accessType...)))
}

func (e *EndpointBase) PrecheckRequestBeforeAuth(request Request, smsMessage *string) error {
	return nil
}

type ResourceEndpoint struct {
	api.ResourceBase
	EndpointBase
}

func (e *ResourceEndpoint) Init(resourceType string, operationName string, selfOp api.Operation, accessType ...access_control.AccessType) {
	e.EndpointBase.Init(operationName, accessType...)
	e.ResourceBase.Init(resourceType)
	e.AddOperation(selfOp)
}

// Base type for API endpoints with empty handlers.
type EndpointNoHandler struct{}

func (e *EndpointNoHandler) HandleRequest(request Request) error {
	return nil
}
