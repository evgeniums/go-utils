package api_server

import "github.com/evgeniums/go-backend-helpers/op_context"

type Request interface {
	op_context.WithCtx
}
