package test_utils

import (
	"net/http"
	"os"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/stretchr/testify/require"
)

func InitAppContextNoDb(t *testing.T, testDir string, config ...string) app_context.Context {
	configFile := utils.OptionalArg(AssetsFilePath(testDir, "test_config.json"), config...)
	if !utils.FileExists(configFile) {
		configFile = AssetsFilePath(testDir, configFile)
	}

	app := app_default.New(nil)
	require.NoErrorf(t, app.Init(configFile), "failed to init application context")

	return app
}

func InitAppContext(t *testing.T, testDir string, config ...string) app_context.Context {
	configFile := utils.OptionalArg(AssetsFilePath(testDir, "test_config.json"), config...)
	if !utils.FileExists(configFile) {
		configFile = AssetsFilePath(testDir, configFile)
	}

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
	app.Db().EnableDebug(true)

	return app
}

func prepareOpContext(ctx op_context.Context, name string) {
	ctx.SetName(name)
	errManager := &generic_error.ErrorManagerBase{}
	errManager.Init(http.StatusBadRequest)
	ctx.SetErrorManager(errManager)
}

func SimpleOpContext(app app_context.Context, name string) op_context.Context {
	ctx := &default_op_context.ContextBase{}
	ctx.Init(app, app.Logger(), app.Db())
	prepareOpContext(ctx, name)

	origin := &default_op_context.Origin{}
	origin.SetType(app.Application())
	origin.SetName(app.AppInstance())
	hostname, _ := os.Hostname()
	origin.SetSource(hostname)
	origin.SetUserType("simple_op_context")
	ctx.SetOrigin(origin)

	return ctx
}

func UserOpContext(app app_context.Context, name string, user auth.User, tenancy ...multitenancy.Tenancy) auth.UserContext {
	ctx := &auth.UserContextBase{}
	ctx.Init(app, app.Logger(), app.Db())
	prepareOpContext(ctx, name)
	ctx.SetAuthUser(user)
	t := utils.OptionalArg(nil, tenancy...)
	if t != nil {
		ctx.SetTenancy(t)
	}
	origin := &default_op_context.Origin{}
	origin.SetType(app.Application())
	origin.SetName(app.AppInstance())
	hostname, _ := os.Hostname()
	origin.SetSource(hostname)
	origin.SetUser(user.Display())
	ctx.SetOrigin(origin)
	return ctx
}
