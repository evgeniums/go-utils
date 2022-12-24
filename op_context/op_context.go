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

	Begin()
	End()

	TraceInMethod(methodName string)
	TraceOutMethod()

	SetGenericError(err *generic_error.Error)
	GetGenericError() *generic_error.Error
}

type WithCtx interface {
	Ctx() Context
}
