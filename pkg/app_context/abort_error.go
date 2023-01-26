package app_context

import (
	"errors"
	"fmt"
)

func AbortError(ctx Context, msg string, err ...error) {
	e := errors.New(msg)
	if len(err) > 0 {
		e = fmt.Errorf("%v: %s", msg, err[0])
		if ctx.Logger() != nil {
			ctx.Logger().ErrorRaw(err[0].Error())
		}
	}
	panic(e)
}

func AbortFatal(ctx Context, msg string, err ...error) {
	if ctx.Logger().CheckFatalStack(ctx.Logger(), msg) {
		panic("AbortFatal")
	}
	AbortError(ctx, msg, err...)
}
