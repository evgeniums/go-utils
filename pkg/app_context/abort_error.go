package app_context

import (
	"fmt"
)

func AbortError(ctx Context, msg string, err error) {
	err = fmt.Errorf("%v: %s", msg, err)
	ctx.Logger().ErrorRaw(err.Error())
	panic(err)
}
