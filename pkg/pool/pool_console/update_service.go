package pool_console

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const UpdateServiceCmd string = "update_service"
const UpdateServiceDescription string = "Update service"

func UpdateService() Handler {
	a := &UpdateServiceHandler{}
	a.Init(UpdateServiceCmd, UpdateServiceDescription)
	return a
}

type UpdateServiceHandler struct {
	HandlerBase
	Service string `long:"service" description:"Short name of the service" required:"true"`
	Field   string `long:"field" description:"Field name" required:"true"`
	Value   string `long:"value" description:"Field value"`
}

func (a *UpdateServiceHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	fields := db.Fields{}
	field := strings.ToLower(a.Field)
	if a.Field == "public_port" || a.Field == "private_port" {
		val, err := utils.StrToUint32(a.Value)
		if err != nil {
			fmt.Println("Value must be unsigned integer")
			return err
		}
		fields[field] = val
	} else {
		fields[field] = a.Value
	}

	err := controller.UpdateService(ctx, a.Service, fields, true)
	if err == nil {
		service, err := controller.FindService(ctx, a.Service, true)
		if err == nil {
			if service != nil {
				fmt.Printf("Updated service:\n\n%s\n\n", utils.DumpPrettyJson(service))
			} else {
				fmt.Println("Service not found")
			}
		}
	}
	return err
}
