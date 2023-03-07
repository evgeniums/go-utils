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

type Handler[Cmd any, Result any] struct {
	Cmd    *Cmd
	Result *Result
}

func (h *Handler[Cmd, Result]) Exec(client Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("Handler.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, h.Cmd, h.Result)
	c.SetError(err)
	return err
}

func NewHandler[Cmd any, Result any](cmd *Cmd, result *Result) *Handler[Cmd, Result] {
	e := &Handler[Cmd, Result]{Cmd: cmd, Result: result}
	return e
}

func NewHandlerCmd[Cmd any](cmd *Cmd) *Handler[Cmd, interface{}] {
	e := &Handler[Cmd, interface{}]{Cmd: cmd}
	return e
}

func NewHandlerResult[Result any](result *Result) *Handler[interface{}, Result] {
	e := &Handler[interface{}, Result]{Result: result}
	return e
}
