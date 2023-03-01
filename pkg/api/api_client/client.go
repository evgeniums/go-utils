package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Client interface {
	Exec(ctx op_context.Context, operation api.Operation, cmd interface{}, response interface{}) error
	Transport() interface{}
}

type ClientOperation interface {
	Exec(client Client, ctx op_context.Context, operation api.Operation) error
}

func MakeOperationHandler(client Client, clientOperation ClientOperation) api.OperationHandler {
	return func(ctx op_context.Context, operation api.Operation) error {
		return clientOperation.Exec(client, ctx, operation)
	}
}

type ServiceClient struct {
	api.ResourceBase
	client Client
}

func (s *ServiceClient) Init(client Client, serviceName string) {
	s.client = client
	s.ResourceBase.Init(serviceName, api.ResourceConfig{Service: true})
}

func (s *ServiceClient) Client() Client {
	return s.client
}

func (s *ServiceClient) ApiClient() Client {
	return s.client
}
