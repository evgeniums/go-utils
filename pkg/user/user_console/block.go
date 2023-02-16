package user_console

import (
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

const BlockCmd string = "block"
const BlockDescription string = "Block access"

func Block[T user.User]() console_tool.Handler[*UserCommands[T]] {
	a := &BlockHandler[T]{}
	a.Init(BlockCmd, BlockDescription)
	return a
}

type BlockHandler[T user.User] struct {
	HandlerBase[T]
	LoginData
}

func (a *BlockHandler[T]) Data() interface{} {
	return &a.LoginData
}

func (a *BlockHandler[T]) Execute(args []string) error {

	ctx, ctrl := a.Context(a.Login)
	defer ctx.Close()

	return ctrl.SetBlocked(ctx, a.Login, true, true)
}
