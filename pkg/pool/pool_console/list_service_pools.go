package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ListServicePoolsCmd string = "list_service_pools"
const ListServicePoolsDescription string = "List all pools that use given service"

func ListServicePools() Handler {
	a := &ListServicePoolsHandler{}
	a.Init(ListServicePoolsCmd, ListServicePoolsDescription)
	return a
}

type ListServicePoolsHandler struct {
	HandlerBase
	Name string `long:"name" description:"Short name of the service" required:"true"`
}

func (a *ListServicePoolsHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	pools, err := controller.GetServiceBindings(ctx, a.Name, true)
	if err == nil {
		fmt.Printf("Pools:\n\n%s\n\n", utils.DumpPrettyJson(pools))
	}
	return err
}
