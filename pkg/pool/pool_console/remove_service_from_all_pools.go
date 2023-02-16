package pool_console

const RemoveServiceFromAllPoolsCmd string = "remove_service_from_all_pools"
const RemoveServiceFromAllPoolsDescription string = "Remove service from all pools"

func RemoveServiceFromAllPools() Handler {
	a := &RemoveServiceFromAllPoolsHandler{}
	a.Init(RemoveServiceFromAllPoolsCmd, RemoveServiceFromAllPoolsDescription)
	return a
}

type RemoveServiceFromAllPoolsData struct {
	Service string `long:"service" description:"Short name of the service" required:"true"`
}

type RemoveServiceFromAllPoolsHandler struct {
	HandlerBase
	RemoveServiceFromAllPoolsData
}

func (a *RemoveServiceFromAllPoolsHandler) Data() interface{} {
	return &a.RemoveServiceFromAllPoolsData
}

func (a *RemoveServiceFromAllPoolsHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.RemoveServiceFromAllPools(ctx, a.Service, true)
	return err
}
