package user_console

import (
	"github.com/evgeniums/go-utils/pkg/console_tool"
	"github.com/evgeniums/go-utils/pkg/user"
)

const EmailCmd string = "email"
const EmailDescription string = "Set email"

func Email[T user.User]() console_tool.Handler[*UserCommands[T]] {
	a := &EmailHandler[T]{}
	a.Init(EmailCmd, EmailDescription)
	return a
}

type EmailHandler[T user.User] struct {
	HandlerBase[T]
	WithEmailData
}

func (a *EmailHandler[T]) Data() interface{} {
	return &a.WithEmailData
}

func (a *EmailHandler[T]) Execute(args []string) error {

	ctx, ctrl, err := a.Context(a.Data(), a.Login)
	if err != nil {
		return err
	}
	defer ctx.Close()

	return ctrl.SetEmail(ctx, a.Login, a.Email, true)
}
