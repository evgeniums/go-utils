package pool_api_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api/tenancy_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api/tenancy_service"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func dbModels() []interface{} {
	r := append([]interface{}{}, admin.DbModels()...)
	r = append(r, pool.DbModels()...)
	r = append(r, customer.DbModels()...)
	r = append(r, multitenancy.DbModels()...)
	return r
}

type testContext struct {
	*api_test.TestContext

	LocalPoolController  pool.PoolController
	LocalCustomerManager *customer.Manager

	LocalTenancyController  multitenancy.TenancyController
	RemoteTenancyController *tenancy_client.TenancyClient
}

func initTest(t *testing.T) *testContext {

	var appWithTenancy *app_with_multitenancy.AppWithMultitenancyBase

	buildApp := func(t *testing.T, buildConfig *app_context.BuildConfig) app_context.Context {
		appWithTenancy = app_with_multitenancy.NewApp(buildConfig, nil)
		return appWithTenancy
	}
	initApp := func(t *testing.T, app app_context.Context, configFile string, args []string, configType ...string) error {
		a, ok := app.(*app_with_multitenancy.AppWithMultitenancyBase)
		require.True(t, ok)
		opCtx, err := a.InitWithArgs(configFile, args, configType...)
		if opCtx != nil {
			opCtx.Close()
		}
		return err
	}
	test_utils.SetAppHandlers(buildApp, initApp)

	ctx := &testContext{}

	ctx.TestContext = api_test.InitTest(t, "tenancy", testDir, dbModels())
	require.NotNil(t, appWithTenancy)
	ctx.LocalPoolController = appWithTenancy.Pools().PoolController()
	ctx.LocalCustomerManager = customer.NewManager()
	ctx.LocalTenancyController = appWithTenancy.Multitenancy().TenancyController()

	tenancyService := tenancy_service.NewTenancyService(ctx.LocalTenancyController)
	api_server.AddServiceToServer(ctx.Server.ApiServer(), tenancyService)

	return ctx
}

func TestInit(t *testing.T) {
	// t.Skip("ok")
	ctx := initTest(t)
	ctx.Close()
}
