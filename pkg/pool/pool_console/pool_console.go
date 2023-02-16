package pool_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolCommands struct {
	console_tool.Commands[*PoolCommands]
	MakeController func(app app_context.Context) pool.PoolController
}

func NewPoolCommands(controllerBuilder ...func(app app_context.Context) pool.PoolController) *PoolCommands {
	p := &PoolCommands{}
	p.Construct(p, "pool", "Manage pools")
	p.MakeController = DefaultPoolController
	if len(controllerBuilder) > 0 {
		p.MakeController = controllerBuilder[0]
	}
	p.LoadHandlers()
	return p
}

func (p *PoolCommands) LoadHandlers() {

	p.AddHandlers(AddPool,
		DeletePool,
		ListPools,
		AddService,
		DeleteService,
		ListServices,
		AddServiceToPool,
		RemoveServiceFromPool,
		RemoveAllServicesFromPool,
		ListPoolServices,
		ListServicePools,
		RemoveServiceFromAllPools,
		ShowPool,
		ShowService,
		UpdatePool,
		UpdateService,
		EnablePool,
		DisablePool,
		EnableService,
		DisableService)
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
	ctrl := b.Group.MakeController(ctx.App())
	return ctx, ctrl
}
