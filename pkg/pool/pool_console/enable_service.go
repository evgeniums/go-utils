package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const EnableServiceCmd string = "enable_service"
const EnableServiceDescription string = "Enable service"

func EnableService() Handler {
	a := &EnableServiceHandler{}
	a.Init(EnableServiceCmd, EnableServiceDescription)
	return a
}

type EnableServiceData struct {
	Service string `long:"pool" description:"Short name of the service" required:"true"`
}

type EnableServiceHandler struct {
	HandlerBase
	EnableServiceData
}

func (a *EnableServiceHandler) Data() interface{} {
	return &a.EnableServiceData
}

func (a *EnableServiceHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	fields := db.Fields{}
	fields["active"] = true

	err = controller.UpdateService(ctx, a.Service, fields, true)
	if err == nil {
		pool, err := controller.FindPool(ctx, a.Service, true)
		if err == nil {
			if pool != nil {
				fmt.Printf("Updated service:\n\n%s\n\n", utils.DumpPrettyJson(pool))
			} else {
				fmt.Println("Pool not found")
			}
		}
	}
	return err
}
