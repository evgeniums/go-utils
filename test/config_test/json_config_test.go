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
	configFile := test_utils.AssetsFilePath(testDir, "main.jsonc")
	app := app_default.New(nil)
	require.NoErrorf(t, app.Init(configFile), "failed to init application context")
	defer app.Close()

	assert.True(t, app.Cfg().IsSet("include"))
	assert.True(t, app.Cfg().IsSet("include_merge"))

	assert.True(t, app.Cfg().IsSet("main_section.array_parameter"))

	assert.True(t, app.Cfg().IsSet("main_section.parameter1"))
	assert.True(t, app.Cfg().IsSet("additional_section.parameter3"))
	assert.True(t, app.Cfg().IsSet("main_section.additional_parameter2"))
	assert.True(t, app.Cfg().IsSet("main_section.override_parameter"))
	assert.True(t, app.Cfg().IsSet("main_section.map_parameter.nested_parameter1"))
	assert.True(t, app.Cfg().IsSet("main_section.map_parameter.nested_parameter2"))

	assert.Equal(t, "value1", app.Cfg().GetString("main_section.parameter1"))
	assert.Equal(t, "value2", app.Cfg().GetString("main_section.additional_parameter2"))
	assert.Equal(t, "value3", app.Cfg().GetString("additional_section.parameter3"))
	assert.Equal(t, "new_value", app.Cfg().GetString("main_section.override_parameter"))

	assert.True(t, app.Cfg().IsSet("main_section.array_parameter.0"))
	assert.True(t, app.Cfg().IsSet("main_section.array_parameter.1"))

	assert.Equal(t, "item1", app.Cfg().GetString("main_section.array_parameter.0"))
	assert.Equal(t, "item2", app.Cfg().GetString("main_section.array_parameter.1"))

	arrayParameter := app.Cfg().GetStringSlice("main_section.array_parameter")
	require.Equal(t, 2, len(arrayParameter))
	assert.Equal(t, "item1", arrayParameter[0])
	assert.Equal(t, "item2", arrayParameter[1])

	assert.Equal(t, "subitem1", app.Cfg().GetString("main_section.subsection.other_array.0"))
	assert.Equal(t, "item2", app.Cfg().GetString("main_section.subsection.other_array.1"))

	arrayParameter = app.Cfg().GetStringSlice("main_section.subsection.other_array")
	require.Equal(t, 2, len(arrayParameter))
	assert.Equal(t, "subitem1", arrayParameter[0])
	assert.Equal(t, "item2", arrayParameter[1])

	assert.True(t, app.Cfg().IsSet("main_section.replacable_section.replacable_parameter"))
	assert.Equal(t, "replacable_value", app.Cfg().GetString("main_section.replacable_section.replacable_parameter"))

	assert.Equal(t, "Hi! /*Hello world*/", app.Cfg().GetString("main_section.with_comments1"))
	assert.Equal(t, "Hi! //Hello world", app.Cfg().GetString("main_section.with_comments2"))
}
