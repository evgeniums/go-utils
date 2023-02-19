package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type List struct {
	cmd    api.Query
	result *tenancy_api.ListTenanciesResponse
}

func (l *List) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("ListTenancies.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, l.cmd, l.result)
	c.SetError(err)
	return err
}

func (t *TenancyClient) List(ctx op_context.Context, filter *db.Filter) ([]*multitenancy.TenancyItem, int64, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyClient.List")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// set query
	cmd := api.NewDbQuery(filter)

	// prepare and exec handler
	handler := &List{
		cmd:    cmd,
		result: &tenancy_api.ListTenanciesResponse{},
	}
	err = t.list.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return nil, 0, err
	}

	// done
	return handler.result.Tenancies, handler.result.Count, nil
}
