package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Operation interface {
	SetResource(resource Resource)
	Resource() Resource

	AccessType() access_control.AccessType
}

type Resource interface {
	Type() string
	Id() string
	SetId(val string)

	HasId() bool
	SetHasId(val bool)

	IsService() bool
	SetSetvice(val bool)

	SetParent(parent Resource)
	Parent() Resource
	AddChild(resource Resource)
	Children() []Resource

	AddOperation(operation Operation)
	Operations() []Operation
	EachOperation(handler func(operation Operation) error) error

	PathPrototype() string
	ActualPath() string
	RebuildPath()

	Chain() []Resource
	ChainResourceId(resourceType string) string
}

type ResourceConfig struct {
	HasId   bool
	Service bool
	Id      string
}

type ResourceBase struct {
	ResourceConfig
	resourceType  string
	pathPrototype string
	actualPath    string
	parent        Resource
	children      []Resource
	operations    []Operation
}

func NewResource(resourceType string, config ...ResourceConfig) *ResourceBase {
	r := &ResourceBase{}
	r.Init(resourceType, config...)
	return r
}

func (r *ResourceBase) Init(resourceType string, config ...ResourceConfig) {
	r.resourceType = resourceType
	r.ResourceConfig = utils.OptionalArg(ResourceConfig{}, config...)
	r.RebuildPath()
}

func (r *ResourceBase) RebuildPath() {

	if r.HasId() {
		r.pathPrototype = utils.ConcatStrings("/", r.Type(), "/:", r.Type(), ":")
		r.actualPath = utils.ConcatStrings("/", r.Type(), "/", r.Id())
	} else {
		r.pathPrototype = utils.ConcatStrings("/", r.Type())
		r.actualPath = r.pathPrototype
	}

	for _, child := range r.children {
		child.RebuildPath()
	}
}

func (r *ResourceBase) PathPrototype() string {
	return r.pathPrototype
}

func (r *ResourceBase) ActualPath() string {
	return r.actualPath
}

func (r *ResourceBase) Type() string {
	return r.resourceType
}

func (r *ResourceBase) Id() string {
	return r.ResourceConfig.Id
}

func (r *ResourceBase) SetId(val string) {
	r.ResourceConfig.Id = val
	if val != "" {
		r.SetHasId(true)
	}
	r.RebuildPath()
}

func (r *ResourceBase) HasId() bool {
	return r.ResourceConfig.HasId
}

func (r *ResourceBase) SetHasId(val bool) {
	r.ResourceConfig.HasId = val
	r.RebuildPath()
}

func (r *ResourceBase) IsService() bool {
	return r.ResourceConfig.Service
}

func (r *ResourceBase) SetSetvice(val bool) {
	r.ResourceConfig.Service = val
	r.RebuildPath()
}

func (r *ResourceBase) SetParent(parent Resource) {
	r.parent = parent
	r.RebuildPath()
}

func (r *ResourceBase) Parent() Resource {
	return r.parent
}

func (r *ResourceBase) AddChild(child Resource) {
	r.children = append(r.children, child)
	child.SetParent(r)
	child.RebuildPath()
}

func (r *ResourceBase) Children() []Resource {
	return r.children
}

func (r *ResourceBase) AddOperation(operation Operation) {
	r.operations = append(r.operations, operation)
	operation.SetResource(r)
}

func (r *ResourceBase) Operations() []Operation {
	return r.operations
}

func (r *ResourceBase) EachOperation(handler func(operation Operation) error) error {

	for _, operation := range r.operations {
		err := handler(operation)
		if err != nil {
			return err
		}
	}

	for _, child := range r.children {
		err := child.EachOperation(handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ResourceBase) Chain() []Resource {
	chain := make([]Resource, 0)

	for parent := r.Parent(); parent != nil; parent = parent.Parent() {
		chain = append([]Resource{parent}, chain...)
	}

	return chain
}

func (r *ResourceBase) ChainResourceId(resourceType string) string {
	if r.Type() == resourceType {
		return r.Id()
	}
	for parent := r.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Type() == resourceType {
			return parent.Id()
		}
	}
	return ""
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
