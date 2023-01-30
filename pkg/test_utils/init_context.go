package test_utils

import (
	"os"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/stretchr/testify/require"
)

func InitAppContext(t *testing.T, testDir string, config ...string) app_context.Context {
	configFile := utils.OptionalArg(AssetsFilePath(testDir, "maindb.json"), config...)

	SetupGormDB(t)
	dbPaths := SqliteDatabasesPath()
	t.Logf("Sqlite DB folder: %s", dbPaths)
	if dbPaths != "" && dbPaths != "/" {
		if utils.FileExists(dbPaths) {
			require.NoErrorf(t, os.RemoveAll(dbPaths), "failed to remove files from db folder")
		}
		require.NoErrorf(t, os.MkdirAll(dbPaths, os.ModePerm), "failed to create db folder")
	}

	app := app_default.New(nil)
	require.NoErrorf(t, app.Init(configFile), "failed to init application context")
	require.NoErrorf(t, app.InitDB("db"), "failed to init database")

	return app
}
