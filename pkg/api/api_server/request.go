package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/parameter"
)

// Interface of request to server API.
type Request interface {
	auth.AuthContext
	parameter.WithParameters

	Server() Server
	Response() Response
	Endpoint() Endpoint
	ResourceIds() map[string]string
}

type RequestBase struct {
	op_context.ContextBase
	auth.SessionBase
	auth.UserContextBase
	endpoint Endpoint
}

func (r *RequestBase) Init(app app_context.Context, log logger.Logger, db db.DB, endpoint Endpoint, fields ...logger.Fields) {
	r.ContextBase.Init(app, log, db, fields...)
	r.endpoint = endpoint
}

func (r *RequestBase) DB() db.DB {
	t := r.GetTenancy()
	if t != nil {
		return t.DB()
	}
	return r.ContextBase.DB()
}

func (r *RequestBase) Endpoint() Endpoint {
	return r.endpoint
}

func FullRequestPath(r Request) string {
	return r.Endpoint().Resource().BuildActualPath(r.ResourceIds())
}

func FullRequestServicePath(r Request) string {
	return r.Endpoint().Resource().BuildActualPath(r.ResourceIds(), true)
}
