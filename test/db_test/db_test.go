package db_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/stretchr/testify/assert"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func assetsFilePath(fileName string) string {
	return filepath.Join(testDir, "assets", fileName)
}

func TestInitDb(t *testing.T) {

	configFile := assetsFilePath("maindb.json")
	t.Logf("Sqlite folder: %s", test_utils.SqliteFolder)

	test_utils.SetupGormDB(t)

	app := app_default.New(nil)
	defer app.Close()
	assert.NoErrorf(t, app.Init(configFile), "failed to init application context")
	assert.NoErrorf(t, app.InitDB("db"), "failed to init database")
}
