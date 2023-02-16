package user_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type UserCommands[T user.User] struct {
	console_tool.Commands[*UserCommands[T]]
	MakeController func(app app_context.Context) user.Users[T]
}

func NewUserCommands[T user.User](groupName string, groupDescription string, controllerBuilder func(app app_context.Context) user.Users[T], defaultHandlers ...bool) *UserCommands[T] {
	p := &UserCommands[T]{}
	p.Construct(p, groupName, groupDescription)
	p.MakeController = controllerBuilder
	if utils.OptionalArg(true, defaultHandlers...) {
		p.LoadDefaultHandlers()
	}
	return p
}

func (p *UserCommands[T]) LoadDefaultHandlers() {

	p.AddHandlers(AddWithPhone[T],
		Password[T],
		Phone[T],
		Email[T],
		Block[T],
		Unblock[T],
		List[T],
	)
}

type HandlerBase[T user.User] struct {
	console_tool.HandlerBase[*UserCommands[T]]
}

func (b *HandlerBase[T]) Context(login ...string) (op_context.Context, user.Users[T]) {
	ctx := b.HandlerBase.Context()
	ctrl := b.Group.MakeController(ctx.App())
	if len(login) != 0 {
		err := ctrl.ValidateLogin(login[0])
		if err != nil {
			panic(fmt.Sprintf("Invalid login format: %s", err))
		}
	}
	return ctx, ctrl
}
