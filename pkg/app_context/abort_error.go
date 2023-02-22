package app_context

import (
	"errors"
	"fmt"
	"os"
)

func AbortError(ctx Context, msg string, err ...error) {
	e := errors.New(msg)
	if len(err) > 0 {
		e = fmt.Errorf("%v: %s", msg, err[0])
		if ctx.Logger() != nil {
			ctx.Logger().ErrorRaw(err[0].Error())
		}
	}
	fmt.Println(e)
	fmt.Println("Failed")
	os.Exit(1)
}

func AbortFatal(ctx Context, msg string, err ...error) {
	if ctx.Logger().CheckFatalStack(ctx.Logger(), msg) {
		fmt.Println("Failed")
		os.Exit(1)
	}
	AbortError(ctx, msg, err...)
}
