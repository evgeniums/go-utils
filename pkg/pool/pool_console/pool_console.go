package pool_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type PoolCommands struct {
	console_tool.Commands[*PoolCommands]
	PoolController pool.PoolController
}

func NewPoolCommands(poolController pool.PoolController) *PoolCommands {
	p := &PoolCommands{}
	p.Construct(p, "pool", "Manage pools")
	p.PoolController = poolController
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

type Handler = console_tool.Handler[*PoolCommands]

type HandlerBase struct {
	console_tool.HandlerBase[*PoolCommands]
}

func (b *HandlerBase) Context(data interface{}) (op_context.Context, pool.PoolController, error) {
	ctx, err := b.HandlerBase.Context(data)
	if err != nil {
		return ctx, nil, err
	}
	return ctx, b.Group.PoolController, nil
}
