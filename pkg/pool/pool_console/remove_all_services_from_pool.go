package pool_console

const RemoveAllServicesFromPoolCmd string = "remove_all_services_from_pool"
const RemoveAllServicesFromPoolDescription string = "Remove all services from pool"

func RemoveAllServicesFromPool() Handler {
	a := &RemoveAllServicesFromPoolHandler{}
	a.Init(RemoveAllServicesFromPoolCmd, RemoveAllServicesFromPoolDescription)
	return a
}

type RemoveAllServicesFromPoolHandler struct {
	HandlerBase
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

func (a *RemoveAllServicesFromPoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.RemoveAllServicesFromPool(ctx, a.Pool, true)
	return err
}
