package user_console

import (
	"encoding/json"
	"fmt"

	"github.com/evgeniums/go-utils/pkg/console_tool"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/user"
)

const ListCmd string = "list"
const ListDescription string = "List records"

func List[T user.User]() console_tool.Handler[*UserCommands[T]] {
	a := &ListHandler[T]{}
	a.Init(ListCmd, ListDescription)
	return a
}

type WithQueryData struct {
	console_tool.QueryData
}

type ListHandler[T user.User] struct {
	HandlerBase[T]
	WithQueryData
}

func (a *ListHandler[T]) Data() interface{} {
	return &a.WithQueryData
}

func (a *ListHandler[T]) Execute(args []string) error {

	ctx, ctrl, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	filter, err := db.ParseQuery(ctx.Db(), a.Query, ctrl.MakeUser(), "")
	if err != nil {
		return fmt.Errorf("failed to parse query: %s", err)
	}

	users, count, err := ctrl.FindUsers(ctx, filter)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(users, "", "   ")
	if err != nil {
		return fmt.Errorf("failed to serialize result: %s", err)
	}
	fmt.Printf("********************\n\n%s\n\nCount %d\n\n********************\n\n", string(b), count)
	return nil
}
