package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type ChangePoolOrDb struct {
	cmd *multitenancy.WithPoolAndDb
}

func (s *ChangePoolOrDb) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("ChangePoolOrDb.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, s.cmd, nil)
	c.SetError(err)
	return err
}

func (t *TenancyClient) ChangePoolOrDb(ctx op_context.Context, id string, poolId string, dbName string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.ChangePoolOrDb")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := &ChangePoolOrDb{
		cmd: &multitenancy.WithPoolAndDb{POOL_ID: poolId, DBNAME: dbName},
	}

	// prepare and exec handler
	op := api.NamedResourceOperation(t.TenancyResource, "pool-db", tenancyId, tenancy_api.ChangePoolOrDb())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil

}
