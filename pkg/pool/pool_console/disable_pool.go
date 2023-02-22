package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const DisablePoolCmd string = "disable_pool"
const DisablePoolDescription string = "Disable pool"

func DisablePool() Handler {
	d := &DisablePoolHandler{}
	d.Init(DisablePoolCmd, DisablePoolDescription)
	return d
}

type DisablePoolData struct {
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

type DisablePoolHandler struct {
	HandlerBase
	DisablePoolData
}

func (d *DisablePoolHandler) Data() interface{} {
	return &d.DisablePoolData
}

func (d *DisablePoolHandler) Execute(args []string) error {

	ctx, controller, err := d.Context(d.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	fields := db.Fields{}
	fields["active"] = false

	err = controller.UpdatePool(ctx, d.Pool, fields, true)
	if err == nil {
		pool, err := controller.FindPool(ctx, d.Pool, true)
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
