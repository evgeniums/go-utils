package op_context

import (
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
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

	Tr(phrase string) string
}

type WithCtx interface {
	Ctx() Context
}
