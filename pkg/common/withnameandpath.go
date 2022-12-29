package common

import "github.com/evgeniums/go-backend-helpers/pkg/utils"

// Interface for type having name and path.
type WithNameAndPath interface {
	Name() string
	Path() string
	FullPath() string

	setParentPath(path string, separator ...string)
}

type WithNameAndPathParent interface {
	WithNameAndPath
	AddChild(child WithNameAndPath)
}

// Base type for types having name and path.
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

// Base type for types having name and path and children.
type WithNameAndPathParentBase struct {
	WithNameAndPathBase
}

func (w *WithNameAndPathParentBase) AddChild(child WithNameAndPath) {
	child.setParentPath(w.FullPath())
}
