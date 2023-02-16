package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const DisablePoolCmd string = "disable_pool"
const DisablePoolDescription string = "Disable pool"

func DisablePool() Handler {
	a := &DisablePoolHandler{}
	a.Init(DisablePoolCmd, DisablePoolDescription)
	return a
}

type DisablePoolHandler struct {
	HandlerBase
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

func (a *DisablePoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	fields := db.Fields{}
	fields["active"] = false

	err := controller.UpdatePool(ctx, a.Pool, fields, true)
	if err == nil {
		pool, err := controller.FindPool(ctx, a.Pool, true)
		if err == nil {
			if pool != nil {
				fmt.Printf("Updated pool:\n\n%s\n\n", utils.DumpPrettyJson(pool))
			} else {
				fmt.Println("Pool not found")
			}
		}
	}
	return err
}
