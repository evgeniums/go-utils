package pool_console

const DeletePoolCmd string = "delete_pool"
const DeletePoolDescription string = "Delete pool"

func DeletePool() Handler {
	a := &DeletePoolHandler{}
	a.Init(DeletePoolCmd, DeletePoolDescription)
	return a
}

type DeletePoolHandler struct {
	HandlerBase
	Pool string `long:"pool" description:"Short name of the pool" required:"true"`
}

func (a *DeletePoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.DeletePool(ctx, a.Pool, true)
	return err
}
