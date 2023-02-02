package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func TestLoadJsonConfig(t *testing.T) {
	configFile := test_utils.AssetsFilePath(testDir, "main.json")
	app := app_default.New(nil)
	require.NoErrorf(t, app.Init(configFile), "failed to init application context")
	defer app.Close()

	assert.True(t, app.Cfg().IsSet("main_section.parameter1"))
	assert.True(t, app.Cfg().IsSet("additional_section.parameter3"))
	assert.True(t, app.Cfg().IsSet("main_section.additional_parameter2"))

	assert.Equal(t, "value1", app.Cfg().GetString("main_section.parameter1"))
	assert.Equal(t, "value2", app.Cfg().GetString("main_section.additional_parameter2"))
	assert.Equal(t, "value3", app.Cfg().GetString("additional_section.parameter3"))
}
