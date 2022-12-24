package api_server

import "github.com/evgeniums/go-backend-helpers/utils"

type WithNameAndPath interface {
	Name() string
	Path() string
	FullPath() string

	setParentPath(path string, separator ...string)
}

type WithNameAndPathBase struct {
	path string
	name string

	fullPath string
}

func (e *WithNameAndPathBase) Init(path string, name string) {
	e.path = path
	e.fullPath = path
	e.name = name
}

func (e *WithNameAndPathBase) Path() string {
	return e.path
}

func (e *WithNameAndPathBase) FullPath() string {
	return e.fullPath
}

func (e *WithNameAndPathBase) Name() string {
	return e.name
}

func (e *WithNameAndPathBase) setParentPath(path string, separator ...string) {
	sep := utils.OptionalArg("/", separator...)
	e.fullPath = path + sep + e.path
}

type Endpoint interface {
	WithNameAndPath

	HandleRequest(request Request)

	Is2FaDefault() bool
	PrecheckRequestBefore2Fa(request Request)
}

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
