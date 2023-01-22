package app_context

import (
	"fmt"
)

func AbortError(ctx Context, msg string, err error) {
	err = fmt.Errorf("%v: %s", msg, err)
	if ctx.Logger() != nil {
		ctx.Logger().ErrorRaw(err.Error())
	}
	panic(err)
}
