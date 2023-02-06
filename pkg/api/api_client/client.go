package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Client interface {
	Exec(ctx op_context.Context, operation api.Operation, cmd interface{}, response interface{}) generic_error.Error
}

type ClientHandler interface {
	Exec(client Client, ctx op_context.Context, operation api.Operation) generic_error.Error
}

func Handler(client Client, clientHandler ClientHandler) api.OperationHandler {
	return func(ctx op_context.Context, operation api.Operation) generic_error.Error {
		return clientHandler.Exec(client, ctx, operation)
	}
}
