package console_tool

import (
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/jessevdk/go-flags"
)

type ControllerBuilder[T any] func(app app_context.Context) T

type CommandGroupI interface {
	Name() string
}

type CommandGroup[T any] struct {
	CommandGroupBase
	MakeController ControllerBuilder[T]
}

func (c *CommandGroup[T]) NewController(app app_context.Context) T {
	return c.MakeController(app)
}

func (c *CommandGroup[T]) Init(groupName string, groupDescription string, controllerBuilder ControllerBuilder[T]) {
	c.GroupName = groupName
	c.GroupDescription = groupDescription
	c.MakeController = controllerBuilder
}

type CommandGroupBase struct {
	GroupName        string
	GroupDescription string

	FillHandlers func(ctxBuilder ContextBulder, group *flags.Command)
}

func (c *CommandGroupBase) Name() string {
	return c.GroupName
}

func (c *CommandGroupBase) Handlers(ctxBuilder ContextBulder, parser *flags.Parser) {

	group, err := parser.AddCommand(c.GroupName, c.GroupDescription, "", &Dummy{})
	if err != nil {
		fmt.Printf("failed to add %s group: %s", c.GroupName, err)
		os.Exit(1)
	}

	c.FillHandlers(ctxBuilder, group)
}

func AddCommand[T CommandGroupI](parent *flags.Command, ctxBuilder ContextBulder, group T, makeHandler func() Handler[T]) {
	handler := makeHandler()
	handler.Construct(ctxBuilder, group)
	parent.AddCommand(handler.HandlerName(), handler.HandlerDescription(), "", handler)
}

type Handler[T CommandGroupI] interface {
	CtxBuilder() ContextBulder
	HandlerGroup() T
	Construct(ctxBuilder ContextBulder, group T)
	HandlerName() string
	HandlerDescription() string
}

type HandlerBaseHolder[T CommandGroupI] struct {
	CtxBuilder  ContextBulder
	Group       T
	Name        string
	Description string
}

type HandlerBase[T CommandGroupI] struct {
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
