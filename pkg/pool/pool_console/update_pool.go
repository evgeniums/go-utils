package pool_console

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const UpdatePoolCmd string = "update_pool"
const UpdatePoolDescription string = "Update pool"

func UpdatePool() poolsHandler {
	a := &UpdatePoolHandler{}
	a.Init(UpdatePoolCmd, UpdatePoolDescription)
	return a
}

type UpdatePoolHandler struct {
	poolsHandlerBase
	Pool  string `long:"pool" description:"Short name of the pool" required:"true"`
	Field string `long:"field" description:"Field name" required:"true"`
	Value string `long:"value" description:"Field value"`
}

func (a *UpdatePoolHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	field := strings.ToLower(a.Field)
	fields := db.Fields{}
	fields[field] = a.Value

	err := controller.UpdatePool(ctx, a.Pool, fields, true)
	if err == nil {
		pool, err := controller.FindPool(ctx, a.Pool, true)
		if err == nil {
			if pool != nil {
				fmt.Printf("Updated pool:\n\n%s\n\n", utils.DumpPrettyJson(pool))
			} else {
				fmt.Println("Pool not found")
			}
		}
	}
	return err
}
