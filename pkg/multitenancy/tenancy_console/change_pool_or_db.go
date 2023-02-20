package tenancy_console

import "github.com/evgeniums/go-backend-helpers/pkg/multitenancy"

const ChangePoolOrDbCmd string = "pool-db"
const ChangePoolOrDbDescription string = "Change tenancy's pool and/or database name"

func ChangePoolOrDb() Handler {
	a := &ChangePoolOrDbHandler{}
	a.Init(ChangePoolOrDbCmd, ChangePoolOrDbDescription)
	return a
}

type ChangePoolOrDbData struct {
	TenancySelector
	multitenancy.WithPoolAndDb
}

type ChangePoolOrDbHandler struct {
	HandlerBase
	ChangePoolOrDbData
}

func (a *ChangePoolOrDbHandler) Data() interface{} {
	return &a.ChangePoolOrDbData
}

func (a *ChangePoolOrDbHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	id, idIsDisplay := PrepareId(a.Id, a.Customer, a.Role)
	return controller.ChangePoolOrDb(ctx, id, a.POOL_ID, a.DBNAME, idIsDisplay)
}
