package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/op_context"

// Interface of response of server API.
type Response interface {
	op_context.WithCtx
	WithAuth
}
