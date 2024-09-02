package tenancy_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/api/api_client"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/op_context"
)

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
	handler := api_client.NewHandler(cmd, &api.ResponseExists{})
	err = t.exists.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return false, err
	}

	// done
	return handler.Result.Exists, nil
}
