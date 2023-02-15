package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ShowServiceCmd string = "show_service"
const ShowServiceDescription string = "Show service"

func ShowService() poolsHandler {
	a := &ShowServiceHandler{}
	a.Init(ShowServiceCmd, ShowServiceDescription)
	return a
}

type ShowServiceHandler struct {
	poolsHandlerBase
	Service string `long:"pool" description:"Short name of the service" required:"true"`
}

func (a *ShowServiceHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	service, err := controller.FindPool(ctx, a.Service, true)
	if err == nil {
		if service != nil {
			fmt.Printf("Service:\n\n%s\n\n", utils.DumpPrettyJson(service))
		} else {
			fmt.Println("Service not found")
		}
	}
	return err
}
