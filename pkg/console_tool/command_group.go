package console_tool

import (
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/jessevdk/go-flags"
)

type CommandGroup interface {
	Name() string
	Description() string
}

type CommandGroupBase struct {
	GroupName        string
	GroupDescription string
}

func NewCommandGroup(groupName string, groupDescription string) *CommandGroupBase {
	c := &CommandGroupBase{}
	c.GroupName = groupName
	c.GroupDescription = groupDescription
	return c
}

func (c *CommandGroupBase) Name() string {
	return c.GroupName
}

func (c *CommandGroupBase) Description() string {
	return c.GroupDescription
}

type HandlerBuilder[T CommandGroup] func() Handler[T]

type Commands[T CommandGroup] struct {
	*CommandGroupBase
	Self          T
	ExtraHandlers []HandlerBuilder[T]
}

func (c *Commands[T]) AddHandlers(handlers ...HandlerBuilder[T]) {
	c.ExtraHandlers = append(c.ExtraHandlers, handlers...)
}

func (c Commands[T]) AddCommand(parent *flags.Command, ctxBuilder ContextBulder, group T) {
	for _, handler := range c.ExtraHandlers {
		AddCommand(parent, ctxBuilder, group, handler)
	}
}

func (c *Commands[T]) Handlers(ctxBuilder ContextBulder, parser *flags.Parser) {

	parent, err := parser.AddCommand(c.Name(), c.Description(), "", &Dummy{})
	if err != nil {
		fmt.Printf("failed to add %s group: %s", c.Name(), err)
		os.Exit(1)
	}

	for _, handler := range c.ExtraHandlers {
		AddCommand(parent, ctxBuilder, c.Self, handler)
	}
}

func (c *Commands[T]) Description() string {
	return c.CommandGroupBase.Description()
}

func (c *Commands[T]) Name() string {
	return c.CommandGroupBase.Name()
}

func (c *Commands[T]) Construct(self T, name string, description string) {
	c.Self = self
	c.CommandGroupBase = NewCommandGroup(name, description)
	c.ExtraHandlers = make([]HandlerBuilder[T], 0)
}

//-------------------------------------------------

func AddCommand[T CommandGroup](parent *flags.Command, ctxBuilder ContextBulder, group T, makeHandler HandlerBuilder[T]) {
	handler := makeHandler()
	handler.Construct(ctxBuilder, group)
	parent.AddCommand(handler.HandlerName(), handler.HandlerDescription(), "", handler.Data())
}

//-------------------------------------------------

type Handler[T CommandGroup] interface {
	CtxBuilder() ContextBulder
	HandlerGroup() T
	Construct(ctxBuilder ContextBulder, group T)
	HandlerName() string
	HandlerDescription() string
	Data() interface{}
}

type HandlerBaseHolder[T CommandGroup] struct {
	CtxBuilder  ContextBulder
	Group       T
	Name        string
	Description string
}

type HandlerBase[T CommandGroup] struct {
	HandlerBaseHolder[T]
}

func (h *HandlerBase[T]) Construct(ctxBuilder ContextBulder, group T) {
	h.HandlerBaseHolder.CtxBuilder = ctxBuilder
	h.HandlerBaseHolder.Group = group
}

func (h *HandlerBase[T]) Init(name string, description string) {
	h.HandlerBaseHolder.Name = name
	h.HandlerBaseHolder.Description = description
}

func (h *HandlerBase[T]) CtxBuilder() ContextBulder {
	return h.HandlerBaseHolder.CtxBuilder
}

func (h *HandlerBase[T]) HandlerGroup() T {
	return h.HandlerBaseHolder.Group
}

func (h *HandlerBase[T]) HandlerName() string {
	return h.HandlerBaseHolder.Name
}

func (h *HandlerBase[T]) HandlerDescription() string {
	return h.HandlerBaseHolder.Description
}

func (h *HandlerBase[T]) Context() op_context.Context {
	ctx := h.HandlerBaseHolder.CtxBuilder(h.Group.Name(), h.HandlerBaseHolder.Name)
	return ctx
}

func (h *HandlerBase[T]) Data() interface{} {
	return &Dummy{}
}
