package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type OperationHandler = func(ctx op_context.Context, operation Operation) (generic_error.Error, error)

type Operation interface {
	SetResource(resource Resource)
	Resource() Resource

	AccessType() access_control.AccessType
	Exec(ctx op_context.Context, handler OperationHandler) (generic_error.Error, error)
}

type OperationBase struct {
	resource   Resource
	accessType access_control.AccessType
}

func (o *OperationBase) Init(accessType access_control.AccessType) {
	o.accessType = accessType
}

func (o *OperationBase) SetResource(resource Resource) {
	o.resource = resource
}

func (o *OperationBase) Resource() Resource {
	return o.resource
}

func (o *OperationBase) AccessType() access_control.AccessType {
	return o.accessType
}

func (o *OperationBase) Exec(ctx op_context.Context, handler OperationHandler) (generic_error.Error, error) {
	return handler(ctx, o)
}
