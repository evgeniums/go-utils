package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ListServicesCmd string = "list_services"
const ListServicesDescription string = "List services"

func ListServices() poolsHandler {
	a := &ListServicesHandler{}
	a.Init(ListServicesCmd, ListServicesDescription)
	return a
}

type ListServicesHandler struct {
	poolsHandlerBase
}

func (a *ListServicesHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	services, _, err := controller.GetServices(ctx, nil)
	if err == nil {
		fmt.Printf("Services:\n\n%s\n\n", utils.DumpPrettyJson(services))
	}
	return err
}
