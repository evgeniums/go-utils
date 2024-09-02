package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/utils"
)

const DisableServiceCmd string = "disable_service"
const DisableServiceDescription string = "Disable service"

func DisableService() Handler {
	a := &DisableServiceHandler{}
	a.Init(DisableServiceCmd, DisableServiceDescription)
	return a
}

type DisableServiceData struct {
	Service string `long:"pool" description:"Short name of the service" required:"true"`
}

type DisableServiceHandler struct {
	HandlerBase
	DisableServiceData
}

func (a *DisableServiceHandler) Data() interface{} {
	return &a.DisableServiceData
}

func (a *DisableServiceHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	s, err := pool.DeactivateService(controller, ctx, a.Service, true)
	if err == nil {
		fmt.Printf("Updated service:\n\n%s\n\n", utils.DumpPrettyJson(s))
	}

	return err
}
