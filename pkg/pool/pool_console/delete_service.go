package pool_console

const DeleteServiceCmd string = "delete_service"
const DeleteServiceDescription string = "Delete service"

func DeleteService() poolsHandler {
	a := &DeleteServiceHandler{}
	a.Init(DeleteServiceCmd, DeleteServiceDescription)
	return a
}

type DeleteServiceHandler struct {
	poolsHandlerBase
	Service string `long:"service" description:"Short name of the service" required:"true"`
}

func (a *DeleteServiceHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	err := controller.DeleteService(ctx, a.Service, true)
	return err
}
