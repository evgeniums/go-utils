package pool_console

const RemoveServiceFromPoolCmd string = "remove_service_from_pool"
const RemoveServiceFromPoolDescription string = "Remove service from pool"

func RemoveServiceFromPool() Handler {
	a := &RemoveServiceFromPoolHandler{}
	a.Init(RemoveServiceFromPoolCmd, RemoveServiceFromPoolDescription)
	return a
}

type RemoveServiceFromPoolHandler struct {
	HandlerBase
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
	Role string `long:"role" description:"Role of the service in the pool" required:"true"`
}

func (a *RemoveServiceFromPoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.RemoveServiceFromPool(ctx, a.Pool, a.Role, true)
	return err
}
