package user_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type List[U user.User] struct {
	cmd    api.Query
	result *user_api.ListResponse[U]
}

func (a *List[U]) Exec(client api_client.Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("List.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, a.cmd, a.result)
	c.SetError(err)
	return err
}

func (u *UserClient[U]) FindUsers(ctx op_context.Context, filter *db.Filter, users *[]U) error {

	// setup
	var err error
	c := ctx.TraceInMethod("UserClient.FindUsers", logger.Fields{"user_type": u.userTypeName})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// set query
	cmd := &api.DbQuery{}
	if filter != nil {
		cmd.SetQuery(filter.ToQueryString())
	}

	// prepare and exec handler
	handler := &List[U]{
		cmd:    cmd,
		result: &user_api.ListResponse[U]{},
	}
	handler.result.Users = users
	err = u.list.Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return err
	}

	// return result
	return nil
}
