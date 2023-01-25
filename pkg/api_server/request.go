package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/parameter"
)

// Interface of request to server API.
type Request interface {
	auth.AuthContext
	parameter.WithParameters

	Server() Server
	Tenancy() multitenancy.Tenancy

	Response() Response
}

type RequestBase struct {
	op_context.ContextBase
	auth.SessionBase
	user    auth.User
	tenancy multitenancy.Tenancy
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

func (r *RequestBase) Tenancy() multitenancy.Tenancy {
	return r.tenancy
}

func (r *RequestBase) SetTenancy(tenancy multitenancy.Tenancy) {
	r.tenancy = tenancy
	r.SetLoggerField("tenancy", tenancy.Name())
	if tenancy.Cache() != nil {
		r.SetCache(tenancy.Cache())
	}
}

func (r *RequestBase) AuthUser() auth.User {
	return r.user
}

func (r *RequestBase) SetAuthUser(user auth.User) {
	r.user = user
}
