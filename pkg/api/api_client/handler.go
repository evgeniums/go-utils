package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

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

type HandlerCmd[Cmd any] struct {
	Cmd *Cmd
}

func (h *HandlerCmd[Cmd]) Exec(client Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("Handler.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, h.Cmd, nil)
	c.SetError(err)
	return err
}

func NewHandlerCmd[Cmd any](cmd *Cmd) *HandlerCmd[Cmd] {
	e := &HandlerCmd[Cmd]{Cmd: cmd}
	return e
}

type HandlerResult[Result any] struct {
	Result *Result
}

func (h *HandlerResult[Result]) Exec(client Client, ctx op_context.Context, operation api.Operation) error {

	c := ctx.TraceInMethod("Handler.Exec")
	defer ctx.TraceOutMethod()

	err := client.Exec(ctx, operation, nil, h.Result)
	c.SetError(err)
	return err
}

func NewHandlerResult[Result any](result *Result) *HandlerResult[Result] {
	e := &HandlerResult[Result]{Result: result}
	return e
}
