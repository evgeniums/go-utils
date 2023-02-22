package customer_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_console"
)

const NameCmd string = "name"
const NameDescription string = "Set customer name"

func Name() console_tool.Handler[*user_console.UserCommands[*customer.Customer]] {
	a := &NameHandler{}
	a.Init(NameCmd, NameDescription)
	return a
}

type NameData struct {
	Name string `long:"name" description:"Name of the customer"`
}

type WithNameData struct {
	user_console.LoginData
	NameData
}

type NameHandler struct {
	user_console.HandlerBase[*customer.Customer]
	WithNameData
}

func (a *NameHandler) Data() interface{} {
	return &a.WithNameData
}

func (a *NameHandler) Execute(args []string) error {

	ctx, ctrl, err := a.Context(a.Data(), a.Login)
	if err != nil {
		return err
	}
	defer ctx.Close()

	customerManager, ok := ctrl.(*customer.Manager)
	if !ok {
		panic("Invalid type of user controller")
	}

	return customerManager.SetName(ctx, a.Login, a.Name, true)
}
