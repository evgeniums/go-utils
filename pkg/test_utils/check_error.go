package test_utils

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
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
