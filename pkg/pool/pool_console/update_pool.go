package pool_console

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/pool"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/evgeniums/go-utils/pkg/validator"
)

const UpdatePoolCmd string = "update_pool"
const UpdatePoolDescription string = "Update pool"

func UpdatePool() Handler {
	a := &UpdatePoolHandler{}
	a.Init(UpdatePoolCmd, UpdatePoolDescription)
	return a
}

type UpdatePoolData struct {
	Pool  string `long:"pool" description:"Short name of the pool" required:"true"`
	Field string `long:"field" description:"Field name" required:"true"`
	Value string `long:"value" description:"Field value"`
}

type UpdatePoolHandler struct {
	HandlerBase
	UpdatePoolData
}

func (a *UpdatePoolHandler) Data() interface{} {
	return &a.UpdatePoolData
}

func (a *UpdatePoolHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	field := strings.ToLower(a.Field)
	fields := db.Fields{}
	fields[field] = a.Value

	vErr := validator.ValidateMap(ctx.App().Validator(), fields, &pool.PoolBaseData{})
	if vErr != nil {
		app_context.ErrorLn("failed to validate fields")
		return vErr.Err
	}

	p, err := controller.UpdatePool(ctx, a.Pool, fields, true)
	if err != nil {
		return err
	}

	fmt.Printf("Updated pool:\n\n%s\n\n", utils.DumpPrettyJson(p))
	return nil
}
