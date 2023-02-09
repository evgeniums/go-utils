package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type OperationHandler = func(ctx op_context.Context, operation Operation) error

type Operation interface {
	Name() string

	SetResource(resource Resource)
	Resource() Resource

	AccessType() access_control.AccessType
	Exec(ctx op_context.Context, handler OperationHandler) error

	TestOnly() bool
	SetTestOnly(val bool)
}

type OperationBase struct {
	name       string
	resource   Resource
	accessType access_control.AccessType
	testOnly   bool
}

func NewOperation(name string, accessType access_control.AccessType, testOnly ...bool) *OperationBase {
	o := &OperationBase{}
	o.Init(name, accessType, testOnly...)
	return o
}

func (o *OperationBase) Init(name string, accessType access_control.AccessType, testOnly ...bool) {
	o.name = name
	o.accessType = accessType
	o.testOnly = utils.OptionalArg(false, testOnly...)
}

func (o *OperationBase) Name() string {
	return o.name
}

func (o *OperationBase) TestOnly() bool {
	return o.testOnly
}

func (o *OperationBase) SetTestOnly(val bool) {
	o.testOnly = val
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

func (o *OperationBase) Exec(ctx op_context.Context, handler OperationHandler) error {

	c := ctx.TraceInMethod("Operation.Exec")
	defer ctx.TraceOutMethod()

	err := handler(ctx, o)
	c.SetError(err)
	return err
}
