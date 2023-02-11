package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Resource interface {
	Host() string
	SetHost(val string)

	Type() string
	Id() string
	SetId(val string, rebuild ...bool)

	HasId() bool
	SetHasId(val bool)

	IsService() bool
	SetSetvice(val bool)
	IsServicePart() bool
	Service() Resource

	SetParent(parent Resource, rebuild ...bool)
	Parent() Resource
	AddChild(resource Resource, rebuild ...bool)
	AddChildren(resources ...Resource)
	Children() []Resource

	AddOperation(operation Operation, getter ...bool)
	AddOperations(operations ...Operation)
	Operations() []Operation
	Getter() Operation
	EachOperation(handler func(operation Operation) error, recursive ...bool) error

	PathPrototype() string
	ActualPath() string
	FullPathPrototype() string
	FullActualPath() string
	ServicePathPrototype() string
	ServiceActualPath() string
	BuildActualPath(actualResourceIds map[string]string, service ...bool) string
	FillActualPaths(actualResourceIds map[string]string)
	ResetIds()

	RebuildPaths()

	Chain() []Resource
	ChainResourceId(resourceType string) string

	Clone(withOperations bool) Resource
	CloneChain(withOperations bool) Resource
}

type ResourceConfig struct {
	Host    string
	HasId   bool
	Service bool
	Id      string
}

type ResourceBase struct {
	ResourceConfig
	resourceType         string
	pathPrototype        string
	actualPath           string
	fullPathPrototype    string
	fullActualPath       string
	servicePathPrototype string
	serviceActualPath    string
	parent               Resource
	children             []Resource
	operations           []Operation
	getter               Operation
}

func NewResource(resourceType string, config ...ResourceConfig) *ResourceBase {
	r := &ResourceBase{}
	r.Init(resourceType, config...)
	return r
}

func (r *ResourceBase) Init(resourceType string, config ...ResourceConfig) {
	r.children = make([]Resource, 0)
	r.operations = make([]Operation, 0)
	r.resourceType = resourceType
	r.ResourceConfig = utils.OptionalArg(ResourceConfig{}, config...)
	r.RebuildPaths()
}

func (r *ResourceBase) RebuildPaths() {

	if r.HasId() {
		r.pathPrototype = utils.ConcatStrings("/:", r.Type())
		r.actualPath = utils.ConcatStrings("/", r.Id())
	} else {
		r.pathPrototype = utils.ConcatStrings("/", r.Type())
		r.actualPath = r.pathPrototype
	}

	r.fullPathPrototype = r.pathPrototype
	r.fullActualPath = r.actualPath

	if r.IsService() {
		r.servicePathPrototype = r.pathPrototype
		r.serviceActualPath = r.actualPath
	}

	parent := r.Parent()
	if parent != nil {
		r.fullPathPrototype = utils.ConcatStrings(parent.FullPathPrototype(), r.pathPrototype)
		r.fullActualPath = utils.ConcatStrings(parent.FullActualPath(), r.actualPath)
		if r.IsServicePart() && !r.IsService() {
			r.servicePathPrototype = utils.ConcatStrings(parent.ServicePathPrototype(), r.pathPrototype)
			r.serviceActualPath = utils.ConcatStrings(parent.ServiceActualPath(), r.actualPath)
		}
	}

	for _, child := range r.children {
		child.RebuildPaths()
	}
}

func (r *ResourceBase) PathPrototype() string {
	return r.pathPrototype
}

func (r *ResourceBase) ActualPath() string {
	return r.actualPath
}

func (r *ResourceBase) FullPathPrototype() string {
	return r.fullPathPrototype
}

func (r *ResourceBase) FullActualPath() string {
	return r.fullActualPath
}

func (r *ResourceBase) ServicePathPrototype() string {
	return r.servicePathPrototype
}

func (r *ResourceBase) Host() string {
	if r.ResourceConfig.Host == "" && r.Parent() != nil {
		return r.Parent().Host()
	}
	return r.ResourceConfig.Host
}

func (r *ResourceBase) SetHost(val string) {
	r.ResourceConfig.Host = val
}

func (r *ResourceBase) ServiceActualPath() string {
	return r.serviceActualPath
}

func (r *ResourceBase) Type() string {
	return r.resourceType
}

func (r *ResourceBase) Id() string {
	return r.ResourceConfig.Id
}

func (r *ResourceBase) SetId(val string, rebuild ...bool) {
	r.ResourceConfig.Id = val
	if val != "" {
		r.SetHasId(true)
	}
	if utils.OptionalArg(true, rebuild...) {
		r.RebuildPaths()
	}
}

func (r *ResourceBase) HasId() bool {
	return r.ResourceConfig.HasId
}

func (r *ResourceBase) SetHasId(val bool) {
	r.ResourceConfig.HasId = val
	r.RebuildPaths()
}

func (r *ResourceBase) IsService() bool {
	return r.ResourceConfig.Service
}

func (r *ResourceBase) IsServicePart() bool {
	if r.IsService() {
		return true
	}
	if r.Parent() != nil {
		return r.parent.IsServicePart()
	}
	return false
}

func (r *ResourceBase) Service() Resource {
	if r.IsService() {
		return r
	}
	if r.Parent() != nil {
		return r.parent.Service()
	}
	return nil
}

func (r *ResourceBase) SetSetvice(val bool) {
	r.ResourceConfig.Service = val
	r.RebuildPaths()
}

func (r *ResourceBase) SetParent(parent Resource, rebuild ...bool) {
	r.parent = parent
	if utils.OptionalArg(true, rebuild...) {
		r.RebuildPaths()
	}
}

func (r *ResourceBase) Parent() Resource {
	return r.parent
}

func (r *ResourceBase) AddChild(child Resource, rebuild ...bool) {
	r.children = append(r.children, child)
	child.SetParent(r, rebuild...)
}

func (r *ResourceBase) AddChildren(resources ...Resource) {
	for _, child := range resources {
		r.AddChild(child)
	}
}

func (r *ResourceBase) Children() []Resource {
	return r.children
}

func (r *ResourceBase) AddOperation(operation Operation, getter ...bool) {
	r.operations = append(r.operations, operation)
	if utils.OptionalArg(false, getter...) {
		r.getter = operation
	}
	operation.SetResource(r)
}

func (r *ResourceBase) AddOperations(operations ...Operation) {
	for _, op := range operations {
		r.AddOperation(op)
	}
}

func (r *ResourceBase) Getter() Operation {
	return r.getter
}

func (r *ResourceBase) Operations() []Operation {
	return r.operations
}

func (r *ResourceBase) EachOperation(handler func(operation Operation) error, recursive ...bool) error {

	for _, operation := range r.operations {
		err := handler(operation)
		if err != nil {
			return err
		}
	}

	if utils.OptionalArg(true, recursive...) {
		for _, child := range r.children {
			err := child.EachOperation(handler)
			if err != nil {
				return err
			}
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

func (r *ResourceBase) BuildActualPath(actualResourceIds map[string]string, service ...bool) string {

	servicePath := utils.OptionalArg(false, service...)
	if servicePath && !r.IsServicePart() {
		return ""
	}

	path := "/"
	if r.HasId() {
		id, ok := actualResourceIds[r.Type()]
		if ok {
			path = utils.ConcatStrings("/", id)
		}
	} else {
		path = utils.ConcatStrings("/", r.Type())
	}

	parent := r.Parent()
	if parent != nil {
		parentPath := parent.BuildActualPath(actualResourceIds, service...)
		if parentPath != "" {
			path = utils.ConcatStrings(parentPath, path)
		}
	}

	return path
}

func (r *ResourceBase) FillActualPaths(actualResourceIds map[string]string) {

	if r.HasId() {
		selfId, ok := actualResourceIds[r.Type()]
		if ok {
			r.SetId(selfId, false)
		}
	}

	parent := r.Parent()
	if parent != nil {
		parent.FillActualPaths(actualResourceIds)
	} else {
		r.RebuildPaths()
	}
}

func (r *ResourceBase) ResetIds() {
	if r.HasId() {
		r.SetId("", false)
	}
	parent := r.Parent()
	if parent != nil {
		parent.ResetIds()
	} else {
		r.RebuildPaths()
	}
}

func (r *ResourceBase) Clone(withOperations bool) Resource {
	resource := NewResource(r.resourceType, r.ResourceConfig)
	if r.HasId() {
		resource.SetId(r.Id())
	}
	if withOperations {
		r.EachOperation(func(operation Operation) error {
			op := NewOperation(operation.Name(), operation.AccessType(), operation.TestOnly())
			resource.AddOperation(op)
			return nil
		})
	}
	return resource
}

func (r *ResourceBase) CloneChain(withOperations bool) Resource {
	resource := NewResource(r.resourceType, r.ResourceConfig)
	if r.HasId() {
		resource.SetId(r.Id())
	}

	var topResource Resource
	topResource = resource
	for parent := r.Parent(); parent != nil; parent = parent.Parent() {
		parentClone := parent.Clone(withOperations)
		parentClone.AddChild(topResource, false)
		topResource = parentClone
	}
	topResource.RebuildPaths()

	return resource
}

func GroupResource(resourceType string) Resource {
	r := NewResource(resourceType)
	return r
}

func NamedResource(resourceType string) Resource {
	cfg := ResourceConfig{
		HasId: true,
	}
	r := NewResource(resourceType, cfg)

	group := GroupResource(resourceType)
	group.AddChild(r)

	return r
}

func PrepareCollectionAndNameResource(typeName string) (serviceName string, collectionResource Resource, objectResource Resource) {

	serviceName = utils.ConcatStrings(typeName, "s")
	objectResource = NamedResource(typeName)
	collectionResource = objectResource.Parent()

	return
}
