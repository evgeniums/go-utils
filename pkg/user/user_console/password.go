package user_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

const PasswordCmd string = "password"
const PasswordDescription string = "Set new password"

func Password[T user.User]() console_tool.Handler[*UserCommands[T]] {
	a := &PasswordHandler[T]{}
	a.Init(PasswordCmd, PasswordDescription)
	return a
}

type PasswordHandler[T user.User] struct {
	HandlerBase[T]
	LoginData
}

func (a *PasswordHandler[T]) Data() interface{} {
	return &a.LoginData
}

func (a *PasswordHandler[T]) Execute(args []string) error {

	ctx, ctrl := a.Context(a.Login)
	defer ctx.Close()

	password := ReadPassword()

	return ctrl.SetPassword(ctx, a.Login, password, true)
}
