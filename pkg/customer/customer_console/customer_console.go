package customer_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_console"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type CustomerCommands struct {
	*user_console.UserCommands[*customer.Customer]
}

func NewCustomerCommands(managerBuilder ...func(app app_context.Context) user.Users[*customer.Customer]) *CustomerCommands {

	controllerBuilder := utils.OptionalArg(DefaultCustomerManager, managerBuilder...)

	a := &CustomerCommands{}
	a.UserCommands = user_console.NewUserCommands("customer", "Manage customers", controllerBuilder, false)

	a.AddHandlers(user_console.AddNoPassword[*customer.Customer],
		user_console.Password[*customer.Customer],
		user_console.Phone[*customer.Customer],
		user_console.Email[*customer.Customer],
		user_console.Block[*customer.Customer],
		user_console.Unblock[*customer.Customer],
		user_console.List[*customer.Customer],
		Name,
		Description,
	)

	return a
}

func DefaultCustomerManager(app app_context.Context) user.Users[*customer.Customer] {
	manager := customer.NewManager()
	manager.Init(app.Validator())
	return manager
}
