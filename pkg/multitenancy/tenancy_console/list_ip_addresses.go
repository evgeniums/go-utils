package tenancy_console

import (
	"fmt"

	"github.com/evgeniums/go-utils/pkg/console_tool"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/multitenancy"
	"github.com/evgeniums/go-utils/pkg/utils"
)

const ListIpAddressesCmd string = "ip-list"
const ListIpAddressesDescription string = "List IP addresses"

func ListIpAddresses() Handler {
	a := &ListIpAddressesHandler{}
	a.Init(ListIpAddressesCmd, ListIpAddressesDescription)
	return a
}

type ListIpAddressesHandler struct {
	HandlerBase
	console_tool.QueryData
}

func (a *ListIpAddressesHandler) Data() interface{} {
	return &a.QueryData
}

func (a *ListIpAddressesHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	filter, err := db.ParseQuery(ctx.Db(), a.Query, &multitenancy.TenancyItem{}, "")
	if err != nil {
		return fmt.Errorf("failed to parse query: %s", err)
	}

	tenancies, count, err := controller.ListIpAddresses(ctx, filter)
	if err == nil {
		fmt.Printf("IP addresses:\n\n%s\n\nTotal count %d\n\n", utils.DumpPrettyJson(tenancies), count)
	}
	return err
}
