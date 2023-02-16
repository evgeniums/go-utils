package customer_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_console"
)

const DescriptionCmd string = "description"
const DescriptionDescription string = "Set customer description"

func Description() console_tool.Handler[*user_console.UserCommands[*customer.Customer]] {
	a := &DescriptionHandler{}
	a.Init(DescriptionCmd, DescriptionDescription)
	return a
}

type DescriptionData struct {
	Description string `long:"description" description:"Description of the customer"`
}

type WithDescriptionData struct {
	user_console.LoginData
	DescriptionData
}

type DescriptionHandler struct {
	user_console.HandlerBase[*customer.Customer]
	WithDescriptionData
}

func (a *DescriptionHandler) Data() interface{} {
	return &a.WithDescriptionData
}

func (a *DescriptionHandler) Execute(args []string) error {

	ctx, ctrl := a.Context(a.Login)
	defer ctx.Close()

	customerManager, ok := ctrl.(*customer.Manager)
	if !ok {
		panic("Invalid type of user controller")
	}

	return customerManager.SetDescription(ctx, a.Login, a.Description, true)
}
