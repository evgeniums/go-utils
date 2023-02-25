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
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_service"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_inmem"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
	"github.com/evgeniums/go-backend-helpers/test/pool_api_test/pool_test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func dbModels() []interface{} {
	return utils.ConcatSlices(admin.DbModels(), pool.DbModels(), customer.DbModels(), multitenancy.DbModels())
}

type testContext struct {
	*pool_test_utils.PoolTestContext

	LocalCustomerManager *customer.Manager

	LocalTenancyController  multitenancy.TenancyController
	RemoteTenancyController *tenancy_client.TenancyClient

	AppWithTenancy *app_with_multitenancy.AppWithMultitenancyBase
}

func initContext(t *testing.T, newDb bool, configPrefix ...string) *testContext {

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
	ctx.PoolTestContext = &pool_test_utils.PoolTestContext{}
	ctx.TestContext = api_test.InitTest(t, utils.OptionalArg("tenancy", configPrefix...), testDir, dbModels(), newDb)
	require.NotNil(t, appWithTenancy)
	ctx.LocalPoolController = appWithTenancy.Pools().PoolController()
	ctx.RemotePoolController = pool_client.NewPoolClient(ctx.RestApiClient)
	ctx.LocalCustomerManager = customer.NewManager()
	ctx.LocalTenancyController = appWithTenancy.Multitenancy().TenancyController()

	poolService := pool_service.NewPoolService(appWithTenancy.Pools().PoolController())
	api_server.AddServiceToServer(ctx.Server.ApiServer(), poolService)

	tenancyService := tenancy_service.NewTenancyService(ctx.LocalTenancyController)
	api_server.AddServiceToServer(ctx.Server.ApiServer(), tenancyService)

	ctx.AppWithTenancy = appWithTenancy

	return ctx
}

func preparePools(t *testing.T, ctx *testContext, names ...string) []pool.Pool {

	pools := make([]pool.Pool, len(names))
	for i, name := range names {
		pool := pool_test_utils.AddPool(t, ctx.PoolTestContext, name)
		pools[i] = pool
	}

	return pools
}

type TenancyServiceConfig struct {
	Name     string
	Type     string
	Provider string
}

func prepareServices(t *testing.T, ctx *testContext, configs ...*TenancyServiceConfig) []pool.PoolService {

	services := make([]pool.PoolService, len(configs))
	for i, config := range configs {

		cfg := pool_test_utils.DefaultServiceConfig(config.Name)
		cfg.PROVIDER = config.Provider
		cfg.TYPE_NAME = config.Type

		service := pool_test_utils.AddService(t, ctx.PoolTestContext, cfg)
		services[i] = service
	}

	return services
}

type TenancyPoolConfig struct {
	PoolName      string
	DbService     TenancyServiceConfig
	PubsubService TenancyServiceConfig
}

func preparePoolServices(t *testing.T, ctx *testContext, config TenancyPoolConfig) pool.Pool {

	pools := preparePools(t, ctx, config.PoolName)
	p := pools[0]

	services := prepareServices(t, ctx, &config.DbService, &config.PubsubService)

	require.NoError(t, ctx.RemotePoolController.AddServiceToPool(ctx.ClientOp, p.GetID(), services[0].GetID(), pool.TypeDatabase))
	require.NoError(t, ctx.RemotePoolController.AddServiceToPool(ctx.ClientOp, p.GetID(), services[1].GetID(), pool.TypePubsub))

	return p
}

func TestInit(t *testing.T) {

	prepareCtx := initContext(t, true)

	pools1 := prepareCtx.AppWithTenancy.Pools()
	require.NotNil(t, pools1)
	require.Empty(t, pools1.Pools())
	selfPool1, err := pools1.SelfPool()
	assert.Error(t, err)
	assert.Nil(t, selfPool1)

	poolConfig1 := TenancyPoolConfig{
		PoolName: "pool1",
	}
	poolConfig1.DbService = TenancyServiceConfig{Name: "database_service1", Type: pool.TypeDatabase, Provider: "sqlite"}
	poolConfig1.PubsubService = TenancyServiceConfig{Name: "pubsub_service1", Type: pool.TypePubsub, Provider: pubsub_inmem.Provider}
	p1 := preparePoolServices(t, prepareCtx, poolConfig1)

	poolConfig2 := TenancyPoolConfig{
		PoolName: "pool2",
	}
	poolConfig2.DbService = TenancyServiceConfig{Name: "database_service2", Type: pool.TypeDatabase, Provider: "sqlite"}
	poolConfig2.PubsubService = TenancyServiceConfig{Name: "pubsub_service2", Type: pool.TypePubsub, Provider: pubsub_inmem.Provider}
	p2 := preparePoolServices(t, prepareCtx, poolConfig2)

	prepareCtx.Close()

	allPoolsCtx := initContext(t, false)

	pools := allPoolsCtx.AppWithTenancy.Pools()
	require.NotNil(t, pools)
	require.NotEmpty(t, pools.Pools())
	require.Equal(t, 2, len(pools.Pools()))
	selfPool2, err := pools.SelfPool()
	assert.Error(t, err)
	assert.Nil(t, selfPool2)

	pool1, err := pools.PoolByName(poolConfig1.PoolName)
	require.NoError(t, err)
	require.NotEmpty(t, pool1)
	pool1, err = pools.Pool(p1.GetID())
	require.NoError(t, err)
	require.NotEmpty(t, pool1)
	nopool, err := pools.PoolByName("unknown_name")
	assert.Error(t, err)
	assert.Empty(t, nopool)
	dbService1, err := pool1.Service(pool.TypeDatabase)
	require.NoError(t, err)
	require.NotEmpty(t, dbService1)
	assert.Equal(t, "database_service1", dbService1.ServiceName)
	pubsubService1, err := pool1.Service(pool.TypePubsub)
	require.NoError(t, err)
	require.NotEmpty(t, pubsubService1)
	assert.Equal(t, "pubsub_service1", pubsubService1.ServiceName)

	pool2, err := pools.PoolByName(poolConfig2.PoolName)
	require.NoError(t, err)
	require.NotEmpty(t, pool2)
	pool2, err = pools.Pool(p2.GetID())
	require.NoError(t, err)
	require.NotEmpty(t, pool2)
	dbService2, err := pool2.Service(pool.TypeDatabase)
	require.NoError(t, err)
	require.NotEmpty(t, dbService2)
	assert.Equal(t, "database_service2", dbService2.ServiceName)
	pubsubService2, err := pool2.Service(pool.TypePubsub)
	require.NoError(t, err)
	require.NotEmpty(t, pubsubService2)
	assert.Equal(t, "pubsub_service2", pubsubService2.ServiceName)

	allPoolsCtx.Close()

	singlePoolCtx := initContext(t, false, "tenancy_single")

	poolsSingle := singlePoolCtx.AppWithTenancy.Pools()
	require.NotNil(t, poolsSingle)
	require.NotEmpty(t, poolsSingle.Pools())
	require.Equal(t, 1, len(poolsSingle.Pools()))
	selfPool, err := poolsSingle.SelfPool()
	require.NoError(t, err)
	require.NotNil(t, selfPool)
	assert.Equal(t, pool2, selfPool)

	singlePoolCtx.Close()
}
