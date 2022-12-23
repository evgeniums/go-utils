package op_context

import (
	"github.com/evgeniums/go-backend-helpers/db"
	"github.com/evgeniums/go-backend-helpers/generic_error"
	"github.com/evgeniums/go-backend-helpers/logger"
)

type Context interface {
	Name() string
	Id() string

	DB() db.DB
	Logger() logger.Logger

	TraceBegin()
	TraceEnd()

	TraceInMethod(name string)
	TraceOutMethod()

	SetGenericError(err *generic_error.Error)
	GetGenericError() *generic_error.Error
}
