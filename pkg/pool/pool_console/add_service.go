package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const AddServiceCmd string = "add_service"
const AddServiceDescription string = "Add service"

func AddService() Handler {
	a := &AddServiceHandler{}
	a.Init(AddServiceCmd, AddServiceDescription)
	return a
}

type AddServiceHandler struct {
	HandlerBase
	pool.ServiceConfigBase
	pool.SecretsBase
	Name        string `long:"name" description:"Short name of the service, must be unique" required:"true"`
	Type        string `long:"type" description:"Service type" required:"true"`
	LongName    string `long:"long-name" description:"Long name of the service"`
	Description string `long:"description" description:"Service description"`
}

func (a *AddServiceHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	s := pool.NewService()
	s.SetName(a.Name)
	s.SetDescription(a.Description)
	s.SetLongName(a.LongName)
	s.SetType(a.Type)
	s.PoolServiceBaseData.ServiceConfigBase = a.ServiceConfigBase
	s.SecretsBase = a.SecretsBase

	addedService, err := controller.AddService(ctx, s)
	if err == nil {
		fmt.Printf("Added service:\n%s\n", utils.DumpPrettyJson(addedService))
	}
	return err
}
