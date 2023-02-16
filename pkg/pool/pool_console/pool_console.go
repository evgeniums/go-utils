package pool_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/jessevdk/go-flags"
)

const PoolGroup string = "pool"
const PoolGroupDescription string = "Manage pools"

type PoolCommands struct {
	console_tool.CommandGroup[pool.PoolController]
}

func NewPoolCommands(controllerBuilder ...console_tool.ControllerBuilder[pool.PoolController]) *PoolCommands {
	p := &PoolCommands{}
	makeController := DefaultPoolController
	if len(controllerBuilder) > 0 {
		makeController = controllerBuilder[0]
	}
	p.Init(PoolGroup, PoolGroupDescription, makeController)
	p.FillHandlers = p.LoadHandlers
	return p
}

var AddCommand = console_tool.AddCommand[*PoolCommands]

func (p *PoolCommands) LoadHandlers(ctxBuilder console_tool.ContextBulder, group *flags.Command) {

	AddCommand(group, ctxBuilder, p, AddPool)
	AddCommand(group, ctxBuilder, p, DeletePool)
	AddCommand(group, ctxBuilder, p, ListPools)
	AddCommand(group, ctxBuilder, p, AddService)
	AddCommand(group, ctxBuilder, p, DeleteService)
	AddCommand(group, ctxBuilder, p, ListServices)
	AddCommand(group, ctxBuilder, p, AddServiceToPool)
	AddCommand(group, ctxBuilder, p, RemoveServiceFromPool)
	AddCommand(group, ctxBuilder, p, RemoveAllServicesFromPool)
	AddCommand(group, ctxBuilder, p, ListPoolServices)
	AddCommand(group, ctxBuilder, p, ListServicePools)
	AddCommand(group, ctxBuilder, p, RemoveServiceFromAllPools)
	AddCommand(group, ctxBuilder, p, ShowPool)
	AddCommand(group, ctxBuilder, p, ShowService)
	AddCommand(group, ctxBuilder, p, UpdatePool)
	AddCommand(group, ctxBuilder, p, UpdateService)
	AddCommand(group, ctxBuilder, p, EnablePool)
	AddCommand(group, ctxBuilder, p, DisablePool)
	AddCommand(group, ctxBuilder, p, EnableService)
	AddCommand(group, ctxBuilder, p, DisableService)
}

func DefaultPoolController(app app_context.Context) pool.PoolController {
	ctrl := pool.NewPoolController(&crud.DbCRUD{})
	return ctrl
}

type Handler = console_tool.Handler[*PoolCommands]

type HandlerBase struct {
	console_tool.HandlerBase[*PoolCommands]
}

func (b *HandlerBase) Context() (op_context.Context, pool.PoolController) {
	ctx := b.HandlerBase.Context()
	ctrl := b.Group.NewController(ctx.App())
	return ctx, ctrl
}
