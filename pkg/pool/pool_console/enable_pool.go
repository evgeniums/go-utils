package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const EnablePoolCmd string = "enable_pool"
const EnablePoolDescription string = "Enable pool"

func EnablePool() Handler {
	a := &EnablePoolHandler{}
	a.Init(EnablePoolCmd, EnablePoolDescription)
	return a
}

type EnablePoolData struct {
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

type EnablePoolHandler struct {
	HandlerBase
	EnablePoolData
}

func (a *EnablePoolHandler) Data() interface{} {
	return &a.EnablePoolData
}

func (a *EnablePoolHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	fields := db.Fields{}
	fields["active"] = true

	p, err := controller.UpdatePool(ctx, a.Pool, fields, true)
	if err == nil {
		fmt.Printf("Updated pool:\n\n%s\n\n", utils.DumpPrettyJson(p))
	}
	return err
}
