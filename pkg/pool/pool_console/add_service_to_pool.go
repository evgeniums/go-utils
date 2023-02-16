package pool_console

const AddServiceToPoolCmd string = "add_service_to_pool"
const AddServiceToPoolDescription string = "Add service to pool"

func AddServiceToPool() Handler {
	a := &AddServiceToPoolHandler{}
	a.Init(AddServiceToPoolCmd, AddServiceToPoolDescription)
	return a
}

type AddServiceToPoolHandler struct {
	HandlerBase
	Pool    string `long:"pool" description:"Short name of the pool" required:"true"`
	Service string `long:"service" description:"Short name of the service" required:"true"`
	Role    string `long:"role" description:"Role of the service in the pool, must be unique per the pool and aphanumeric" required:"true"`
}

func (a *AddServiceToPoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.AddServiceToPool(ctx, a.Pool, a.Service, a.Role, true)
	return err
}
