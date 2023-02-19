package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Exists struct {
	cmd    api.Query
	result *api.ResponseExists
}

func (e *Exists) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("TenancyExists.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, e.cmd, e.result)
	c.SetError(err)
	return err
}

func (t *TenancyClient) Exists(ctx op_context.Context, fields db.Fields) (bool, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyClient.Exists")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// set query
	filter := db.NewFilter()
	filter.Fields = fields
	cmd := api.NewDbQuery(filter)

	// prepare and exec handler
	handler := &Exists{
		cmd:    cmd,
		result: &api.ResponseExists{},
	}
	err = t.exists.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return false, err
	}

	// done
	return handler.result.Exists, nil
}
