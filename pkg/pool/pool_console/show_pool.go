package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ShowPoolCmd string = "show_pool"
const ShowPoolDescription string = "Show pool"

func ShowPool() poolsHandler {
	a := &ShowPoolHandler{}
	a.Init(ShowPoolCmd, ShowPoolDescription)
	return a
}

type ShowPoolHandler struct {
	poolsHandlerBase
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

func (a *ShowPoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	pool, err := controller.FindPool(ctx, a.Pool, true)
	if err == nil {
		if pool != nil {
			fmt.Printf("Pool:\n\n%s\n\n", utils.DumpPrettyJson(pool))
		} else {
			fmt.Println("Pool not found")
		}
	}
	return err
}
