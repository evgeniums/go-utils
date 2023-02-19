package pool_api_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api/tenancy_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api/tenancy_service"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDbModels(t, app, admin.DbModels())
	test_utils.CreateDbModels(t, app, pool.DbModels())
	test_utils.CreateDbModels(t, app, customer.DbModels())
	test_utils.CreateDbModels(t, app, multitenancy.DbModels())
}

type testContext struct {
	*api_test.TestContext

	LocalPoolController  pool.PoolController
	LocalCustomerManager *customer.Manager

	LocalTenancyController  multitenancy.TenancyController
	RemoteTenancyController *tenancy_client.TenancyClient
}

func initTest(t *testing.T) *testContext {

	ctx := &testContext{}

	ctx.TestContext = api_test.InitTest(t, "tenancy", testDir, createDb)
	ctx.LocalPoolController = pool.NewPoolController(&crud.DbCRUD{})
	ctx.LocalCustomerManager = customer.NewManager()

	// TODO fix tenancy manager
	tenancyManager := &tenancy_manager.TenancyManager{}
	ctx.LocalTenancyController = tenancy_manager.NewTenancyController(&crud.DbCRUD{}, tenancyManager)

	tenancyService := tenancy_service.NewTenancyService(ctx.LocalTenancyController)
	api_server.AddServiceToServer(ctx.Server.ApiServer(), tenancyService)

	return ctx
}

func TestInit(t *testing.T) {
	ctx := initTest(t)
	ctx.Close()
}
