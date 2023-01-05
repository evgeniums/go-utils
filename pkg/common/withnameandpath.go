package common

import (
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type WithName interface {
	Name() string
}

func ConstructPath(sections []string, separator string) string {
	return strings.Join(sections, separator)
}

// Interface for type having name and path.
type WithPath interface {
	Path() string
	FullPath() string
	Paths() []string
	Sections() []string
	Separator() string

	SetParent(parent WithPath)
}

type WithNameAndPath interface {
	WithName
	WithPath
}

type WithPathParent interface {
	WithPath
	AddChild(child WithPath)
}

type WithNameBase struct {
	name string
}

func (e *WithNameBase) Init(name string) {
	e.name = name
}

func (w *WithNameBase) Name() string {
	return w.name
}

// Base type for types having name and path.
type WithPathBase struct {
	path      string
	sections  []string
	separator string
}

func (w *WithPathBase) Init(path string, separator ...string) {
	w.path = path
	w.separator = utils.OptionalArg("/", separator...)
	w.sections = strings.Split(w.path, w.separator)
}

func (w *WithPathBase) Path() string {
	return w.path
}

func (w *WithPathBase) Separator() string {
	return w.separator
}

func (w *WithPathBase) Sections() []string {
	return w.sections
}

func (w *WithPathBase) Paths() []string {
	paths := []string{"/"}
	for i := 1; i < len(w.sections); i++ {
		paths = append(paths, ConstructPath(w.sections[0:i+1], w.separator))
	}
	return paths
}

func (w *WithPathBase) FullPath() string {
	return ConstructPath(w.sections, w.separator)
}

func (w *WithPathBase) SetParent(parent WithPath) {
	w.sections = strings.Split(w.path, w.separator)
	ps := parent.Sections()
	last := ps[len(ps)-1]
	if last == "" {
		w.sections = append(ps[:len(ps)-1], w.sections[1:]...)
	} else {
		w.sections = append(ps, w.sections[1:]...)
	}
}

// Base type for types having name and path and children.
type WithPathParentBase struct {
	WithPathBase
}

func (w *WithPathParentBase) AddChild(child WithPath) {
	child.SetParent(w)
}

// Base type for types having name and path.
type WithNameAndPathBase struct {
	WithNameBase
	WithPathBase
}

func (e *WithNameAndPathBase) Init(path string, name string, separator ...string) {
	e.WithNameBase.Init(name)
	e.WithPathBase.Init(path, separator...)
}

// Base type for types having name and path and children.
type WithNameAndPathParentBase struct {
	WithNameBase
	WithPathParentBase
}
