package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ListPoolsCmd string = "list_pools"
const ListPoolsDescription string = "List pools"

func ListPools() poolsHandler {
	a := &ListPoolsHandler{}
	a.Init(ListPoolsCmd, ListPoolsDescription)
	return a
}

type ListPoolsHandler struct {
	poolsHandlerBase
}

func (a *ListPoolsHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	pools, _, err := controller.GetPools(ctx, nil)
	if err == nil {
		fmt.Printf("Pools:\n\n%s\n\n", utils.DumpPrettyJson(pools))
	}
	return err
}
