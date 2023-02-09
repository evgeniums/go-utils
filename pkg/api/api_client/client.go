package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Client interface {
	Exec(ctx op_context.Context, operation api.Operation, cmd interface{}, response interface{}) generic_error.Error
}

type ClientOperation interface {
	Exec(client Client, ctx op_context.Context, operation api.Operation) generic_error.Error
}

func MakeOperationHandler(client Client, clientOperation ClientOperation) api.OperationHandler {
	return func(ctx op_context.Context, operation api.Operation) generic_error.Error {
		return clientOperation.Exec(client, ctx, operation)
	}
}

type ServiceClient struct {
	api.ResourceBase
	Client Client
}

func (s *ServiceClient) Init(client Client, pathName string) {
	s.Client = client
	s.ResourceBase.Init(pathName, api.ResourceConfig{Service: true})
}
