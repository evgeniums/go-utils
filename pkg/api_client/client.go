package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Client interface {
	Exec(ctx op_context.Context, operation Operation, cmd interface{}, response interface{}) (generic_error.Error, error)
}
