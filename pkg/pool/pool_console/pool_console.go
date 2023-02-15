package pool_console

import (
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/jessevdk/go-flags"
)

const PoolGroup string = "pool"

func (a *PoolCommands) FillHandlers(ctxBuilder console_tool.ContextBulder, group *flags.Command) {

	AddCommand(group, ctxBuilder, a, AddPool)
	AddCommand(group, ctxBuilder, a, DeletePool)
	AddCommand(group, ctxBuilder, a, ListPools)
	AddCommand(group, ctxBuilder, a, AddService)
	AddCommand(group, ctxBuilder, a, DeleteService)
	AddCommand(group, ctxBuilder, a, ListServices)
	AddCommand(group, ctxBuilder, a, AddServiceToPool)
	AddCommand(group, ctxBuilder, a, RemoveServiceFromPool)
	AddCommand(group, ctxBuilder, a, RemoveAllServicesFromPool)
	AddCommand(group, ctxBuilder, a, ListPoolServices)
	AddCommand(group, ctxBuilder, a, ListServicePools)
	AddCommand(group, ctxBuilder, a, RemoveServiceFromAllPools)
	AddCommand(group, ctxBuilder, a, ShowPool)
	AddCommand(group, ctxBuilder, a, ShowService)
	AddCommand(group, ctxBuilder, a, UpdatePool)
	AddCommand(group, ctxBuilder, a, UpdateService)
	AddCommand(group, ctxBuilder, a, EnablePool)
	AddCommand(group, ctxBuilder, a, DisablePool)
	AddCommand(group, ctxBuilder, a, EnableService)
	AddCommand(group, ctxBuilder, a, DisableService)
}

func (a *PoolCommands) Handlers(ctxBuilder console_tool.ContextBulder, parser *flags.Parser) {

	pools, err := parser.AddCommand(PoolGroup, "Manage pools", "", &console_tool.Dummy{})
	if err != nil {
		fmt.Printf("failed to add pool group: %s", err)
		os.Exit(1)
	}

	a.FillHandlers(ctxBuilder, pools)
}

func AddCommand(parent *flags.Command, ctxBuilder console_tool.ContextBulder, group *PoolCommands, makeHandler func() poolsHandler) {
	handler := makeHandler()
	handler.Construct(ctxBuilder, group)
	parent.AddCommand(handler.HandlerName(), handler.HandlerDescription(), "", handler)
}

type PoolControllerBuilder = func(app app_context.Context) pool.PoolController

func DefaultPoolController(app app_context.Context) pool.PoolController {
	ctrl := pool.NewPoolController(&crud.DbCRUD{})
	return ctrl
}

type PoolCommands struct {
	MakeController PoolControllerBuilder
}

func NewPoolCommands(controllerBuilder ...PoolControllerBuilder) *PoolCommands {
	p := &PoolCommands{}
	p.MakeController = utils.OptionalArg(DefaultPoolController, controllerBuilder...)
	return p
}

func (a *PoolCommands) NewController(app app_context.Context) pool.PoolController {
	ctrl := a.MakeController(app)
	return ctrl
}

type poolsHandler interface {
	CtxBuilder() console_tool.ContextBulder
	HandlerGroup() *PoolCommands
	Construct(ctxBuilder console_tool.ContextBulder, group *PoolCommands)
	HandlerName() string
	HandlerDescription() string
}

type poolsHandlerBase struct {
	ctxBuilder  console_tool.ContextBulder
	group       *PoolCommands
	name        string
	description string
}

func (b *poolsHandlerBase) Construct(ctxBuilder console_tool.ContextBulder, group *PoolCommands) {
	b.ctxBuilder = ctxBuilder
	b.group = group
}

func (b *poolsHandlerBase) Init(name string, description string) {
	b.name = name
	b.description = description
}

func (b *poolsHandlerBase) CtxBuilder() console_tool.ContextBulder {
	return b.ctxBuilder
}

func (b *poolsHandlerBase) HandlerGroup() *PoolCommands {
	return b.group
}

func (b *poolsHandlerBase) HandlerName() string {
	return b.name
}

func (b *poolsHandlerBase) HandlerDescription() string {
	return b.description
}

func (b *poolsHandlerBase) Context() (op_context.Context, pool.PoolController) {
	ctx := b.ctxBuilder(PoolGroup, b.name)
	ctrl := b.group.NewController(ctx.App())
	return ctx, ctrl
}
