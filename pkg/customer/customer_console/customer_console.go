package customer_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_console"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Commands[T customer.User] struct {
	*user_console.UserCommands[T]
}

type Config[T customer.User] struct {
	ManagerBuilder func(app app_context.Context) user.Users[T]
	Name           string
	Description    string
}

func NewCommands[T customer.User](config Config[T]) *Commands[T] {

	a := &Commands[T]{}
	a.UserCommands = user_console.NewUserCommands(config.Name, config.Description, config.ManagerBuilder, false)

	a.AddHandlers(user_console.AddNoPassword[T],
		user_console.Password[T],
		user_console.Phone[T],
		user_console.Email[T],
		user_console.Block[T],
		user_console.Unblock[T],
		user_console.List[T],
		Name[T],
		Description[T],
	)

	return a
}

type CustomerCommands = Commands[*customer.Customer]

func NewCustomerCommands(managerBuilder ...func(app app_context.Context) user.Users[*customer.Customer]) *CustomerCommands {

	config := Config[*customer.Customer]{
		Name:           "Customer",
		Description:    "Manage customers",
		ManagerBuilder: utils.OptionalArg(DefaultCustomerManager, managerBuilder...),
	}

	return NewCommands(config)
}

func DefaultCustomerManager(app app_context.Context) user.Users[*customer.Customer] {
	manager := customer.NewManager()
	manager.Init(app.Validator())
	return manager
}
