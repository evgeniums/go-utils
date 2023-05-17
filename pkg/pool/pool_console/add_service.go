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

type AddServiceData struct {
	pool.ServiceConfigBase
	pool.SecretsBase
	Name        string `long:"name" description:"Short name of the service, must be unique" required:"true"`
	Type        string `long:"type" description:"Service type" required:"true"`
	LongName    string `long:"long-name" description:"Long name of the service"`
	Description string `long:"description" description:"Service description"`
	Active      string `long:"active" description:"Service is active" default:"true"`
}

type AddServiceHandler struct {
	HandlerBase
	AddServiceData
}

func (a *AddServiceHandler) Data() interface{} {
	return &a.AddServiceData
}

func (a *AddServiceHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()
	s := pool.NewService()
	s.SetName(a.Name)
	s.SetDescription(a.Description)
	s.SetLongName(a.LongName)
	s.SetTypeName(a.Type)
	s.PoolServiceBaseData.ServiceConfigBase = a.ServiceConfigBase
	s.SecretsBase = a.SecretsBase
	s.SetActive(a.Active == "true")

	// TODO check secret1 and secret2
	addedService, err := controller.AddService(ctx, s)
	if err == nil {
		fmt.Printf("Added service:\n%s\n", utils.DumpPrettyJson(addedService))
	}
	return err
}
