package test_utils

import (
	"encoding/json"
	"testing"

	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/common"
	"github.com/evgeniums/go-utils/pkg/generic_error"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CheckGenericError(t *testing.T, err error, expectedCode string, expectedMessage ...string) {
	assert.Error(t, err)
	gErr, ok := err.(generic_error.Error)
	require.True(t, ok)
	assert.Equal(t, expectedCode, gErr.Code())
	if len(expectedMessage) != 0 {
		assert.Equal(t, expectedMessage[0], gErr.Message())
	}
}

func DumpError(t *testing.T, err error, message ...string) {
	if err == nil {
		t.Logf(utils.OptionalArg("Dump error", message...) + ": no error")
		return
	}
	gErr, ok := err.(generic_error.Error)
	if ok {
		DumpObject(t, gErr, message...)
	} else {
		t.Logf("%s: %s", utils.OptionalArg("Error", message...), err)
	}
}

func ObjectEqual(t *testing.T, left common.Object, right common.Object) {
	if left.GetCreatedAt().Equal(right.GetCreatedAt()) {
		right.SetCreatedAt(left.GetCreatedAt())
	}
	if left.GetUpdatedAt().Equal(right.GetUpdatedAt()) {
		right.SetUpDatedAt(left.GetUpdatedAt())
	}
	assert.Equal(t, left, right)
}

func NoError(t *testing.T, ctx op_context.Context, err error) {
	if err == nil {
		return
	}
	if ctx != nil {
		ctx.Close()
		gErr := ctx.GenericError()
		if gErr != nil {
			DumpObject(t, gErr, "Generic error")
		}
	}
	require.NoError(t, err)
}

func NoErrorApp(t *testing.T, app app_context.Context, err error) {
	if err != nil && app != nil {
		app.Logger().CheckFatalStack(app.Logger())
	}
	require.NoError(t, err)
}

func DumpObject(t *testing.T, obj interface{}, message ...string) {
	result, err := json.MarshalIndent(obj, " ", " ")
	require.NoError(t, err)
	msg := utils.OptionalString("", message...)
	if msg != "" {
		t.Logf("%s:\n%s\n", msg, string(result))
	} else {
		t.Logf("\n%s\n", string(result))
	}
}
