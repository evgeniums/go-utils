package admin_console

import (
	"github.com/evgeniums/go-utils/pkg/admin"
	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/user"
	"github.com/evgeniums/go-utils/pkg/user/user_console"
	"github.com/evgeniums/go-utils/pkg/utils"
)

type AdminCommands struct {
	*user_console.UserCommands[*admin.Admin]
}

func NewAdminCommands(managerBuilder ...func(app app_context.Context) user.Users[*admin.Admin]) *AdminCommands {

	controllerBuilder := utils.OptionalArg(DefaultAdminManager, managerBuilder...)

	a := &AdminCommands{}
	a.UserCommands = user_console.NewUserCommands("administrator", "Manage administrators", controllerBuilder)

	return a
}

func DefaultAdminManager(app app_context.Context) user.Users[*admin.Admin] {
	manager := admin.NewManager()
	manager.Init(app.Validator())
	return manager
}
