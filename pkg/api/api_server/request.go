package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/parameter"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

// Interface of request to server API.
type Request interface {
	auth.AuthContext
	parameter.WithParameters

	Server() Server
	Response() Response
	Endpoint() Endpoint
	ResourceIds() map[string]string

	GetResourceId(resourceType string) string

	ParseVerify(cmd interface{}) error
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

func ParseDbQuery(request Request, models []interface{}, q api.Query, queryName string) (*db.Filter, error) {
	c := request.TraceInMethod("ParseDbQuery", logger.Fields{"query_name": queryName, "query": q.Query()})
	defer request.TraceOutMethod()

	err := request.ParseVerify(q)
	if err != nil {
		return nil, c.SetError(err)
	}

	filter, err := db.ParseQuery(request.DB(), q.Query(), models, queryName, db.EmptyFilterValidator(request.App().Validator()))
	if err != nil {
		vErr, ok := err.(*validator.ValidationError)
		if ok {
			request.SetGenericError(vErr.GenericError(), true)
		}
		c.SetMessage("failed to parse/validate db query")
		return nil, c.SetError(err)
	}

	return filter, nil
}
