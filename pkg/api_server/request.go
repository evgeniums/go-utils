package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

// Interface of request to server API.
type Request interface {
	op_context.WithCtx

	Server() Server
	Tenancy() Tenancy

	WithAuth
	message.WithMessage

	Response() Response
}

type RequestBase struct {
	op_context.ContextBase

	tenancy Tenancy
}

func (r *RequestBase) Init(app app_context.Context, log logger.Logger, db db.DB, fields ...logger.Fields) {
	r.ContextBase.Init(app, log, db, fields...)
}

func (r *RequestBase) DB() db.DB {
	if r.tenancy != nil {
		return r.tenancy.DB()
	}
	return r.ContextBase.DB()
}

func (r *RequestBase) Tenancy() Tenancy {
	return r.tenancy
}

func (r *RequestBase) SetTenancy(tenancy Tenancy) {
	r.tenancy = tenancy
	r.SetLoggerField("tenancy", tenancy.Name())
}
