package pool_console

const RemoveServiceFromAllPoolsCmd string = "remove_service_from_all_pools"
const RemoveServiceFromAllPoolsDescription string = "Remove service from all pools"

func RemoveServiceFromAllPools() poolsHandler {
	a := &RemoveServiceFromAllPoolsHandler{}
	a.Init(RemoveServiceFromAllPoolsCmd, RemoveServiceFromAllPoolsDescription)
	return a
}

type RemoveServiceFromAllPoolsHandler struct {
	poolsHandlerBase
	Service string `long:"service" description:"Short name of the service" required:"true"`
}

func (a *RemoveServiceFromAllPoolsHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.RemoveServiceFromAllPools(ctx, a.Service, true)
	return err
}
