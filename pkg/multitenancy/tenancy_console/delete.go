package tenancy_console

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

const DeleteCmd string = "delete"
const DeleteDescription string = "Delete tenancy"

func Delete() Handler {
	a := &DeleteHandler{}
	a.Init(DeleteCmd, DeleteDescription)
	return a
}

type DeleteData struct {
	TenancySelector
	WithDb bool `long:"with-database" description:"Delete tenancy's database. ATTENTION! Deleted data can not be recovered later!"`
}

type DeleteHandler struct {
	HandlerBase
	DeleteData
}

func (a *DeleteHandler) Data() interface{} {
	return &a.DeleteData
}

func (a *DeleteHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Type YES to confirm operation: ")
	text, _ := reader.ReadString('\n')
	if text == "YES" {
		id, idIsDisplay := PrepareId(a.Id, a.Customer, a.Role)
		return controller.Delete(ctx, id, a.WithDb, idIsDisplay)
	}

	return errors.New("operation cancelled")
}
