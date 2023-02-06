package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Resource interface {
	Host() string

	Type() string
	Id() string
	SetId(val string)

	HasId() bool
	SetHasId(val bool)

	IsService() bool
	SetSetvice(val bool)
	IsServicePart() bool
	Service() Resource

	SetParent(parent Resource)
	Parent() Resource
	AddChild(resource Resource)
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

	RebuildPaths()

	Chain() []Resource
	ChainResourceId(resourceType string) string

	FillHateoasLinks(actualResourceIds map[string]string, links *HateoasLinks)

	SetHateoasLinks(links []*HateoasLink)
	AppendHateoasLink(link *HateoasLink)
	HateoasLinks() []*HateoasLink
	ResetHateoasLinks()
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

	links []*HateoasLink
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
		r.pathPrototype = utils.ConcatStrings("/", r.Type(), "/:", r.Type(), ":")
		r.actualPath = utils.ConcatStrings("/", r.Type(), "/", r.Id())
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

func (r *ResourceBase) SetId(val string) {
	r.ResourceConfig.Id = val
	if val != "" {
		r.SetHasId(true)
	}
	r.RebuildPaths()
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

func (r *ResourceBase) SetParent(parent Resource) {
	r.parent = parent
	r.RebuildPaths()
}

func (r *ResourceBase) Parent() Resource {
	return r.parent
}

func (r *ResourceBase) AddChild(child Resource) {
	r.children = append(r.children, child)
	child.SetParent(r)
	child.RebuildPaths()
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

	var path string
	id, ok := actualResourceIds[r.Type()]
	if ok {
		path = utils.ConcatStrings("/", r.Type(), "/", id)
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

func (r *ResourceBase) FillHateoasLinks(actualResourceIds map[string]string, links *HateoasLinks) {

	links.Links = make([]*HateoasLink, 0)

	addLink := func(host string, target string, operation Operation) {

		if operation == nil {
			return
		}

		link := &HateoasLink{}
		link.Host = host
		link.HttpMethod = access_control.Access2HttpMethod(operation.AccessType())
		link.Path = r.BuildActualPath(actualResourceIds)
		link.Operation = operation.Name()
		link.Target = target

		links.Links = append(links.Links, link)
	}

	// add self operations
	selfHost := r.Host()
	r.EachOperation(func(operation Operation) error { addLink(selfHost, TargetSelf, operation); return nil }, false)

	addGetter := func(target string, resource Resource) {
		addLink(resource.Host(), target, resource.Getter())
	}

	// add children getters
	for _, child := range r.children {
		addGetter(TargetChild, child)
	}
	// add parent getter
	parent := r.Parent()
	if parent != nil {
		addGetter(TargetParent, parent)
	}
}

func (r *ResourceBase) SetHateoasLinks(links []*HateoasLink) {
	r.links = links
	parent := r.Parent()
	if parent != nil {
		for _, link := range links {
			if link.Target == TargetParent {
				parent.AppendHateoasLink(link)
			}
		}
	}
}

func (r *ResourceBase) AppendHateoasLink(link *HateoasLink) {
	if r.links == nil {
		r.links = make([]*HateoasLink, 0, 1)
	}
	r.links = append(r.links, link)
}

func (r *ResourceBase) ResetHateoasLinks() {
	r.links = nil
}

func (r *ResourceBase) HateoasLinks() []*HateoasLink {
	return r.links
}
