package api_server

import "github.com/evgeniums/go-backend-helpers/op_context"

// Interface of request to server API.
type Request interface {
	op_context.WithCtx
	WithAuth
}
