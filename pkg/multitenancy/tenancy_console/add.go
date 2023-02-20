package tenancy_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const AddCmd string = "add"
const AddDescription string = "Add tenancy"

func Add() Handler {
	a := &AddHandler{}
	a.Init(AddCmd, AddDescription)
	return a
}

type AddHandler struct {
	HandlerBase
	multitenancy.TenancyData
}

func (a *AddHandler) Data() interface{} {
	return &a.TenancyData
}

func (a *AddHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	added, err := controller.Add(ctx, &a.TenancyData)
	if err == nil {
		fmt.Printf("Added tenancy:\n%s\n", utils.DumpPrettyJson(added))
	}
	return err
}
