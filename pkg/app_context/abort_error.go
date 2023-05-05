package app_context

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var Testing *testing.T

func testError(err error) {
	if Testing != nil {
		require.NoError(Testing, err)
	}
}

func Panic(msg string, err error) {
	e := fmt.Errorf("%s: %s", msg, err)
	testError(e)
	panic(e)
}

func AbortError(ctx Context, msg string, err ...error) {
	e := errors.New(msg)
	if len(err) > 0 {
		e = fmt.Errorf("%v: %s", msg, err[0])
		if ctx.Logger() != nil {
			ctx.Logger().ErrorRaw(err[0].Error())
		}
	}
	testError(e)
	ErrorLn(e.Error())
	ErrorLn("Failed")
	os.Exit(1)
}

func AbortFatal(ctx Context, msg string, err ...error) {
	if ctx.Logger().CheckFatalStack(ctx.Logger(), msg) {
		testError(errors.New("abort fatal"))
		ErrorLn("Failed")
		os.Exit(1)
	}
	AbortError(ctx, msg, err...)
}

func ErrorLn(msg string) {
	l := log.New(os.Stderr, "", 0)
	l.Println(msg)
}
