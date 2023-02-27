package tenancy_api_test

import (
	"encoding/json"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPostgresConfig() DbConfig {
	cfg := DbConfig{}
	cfg.DbHost = "127.0.0.1"
	cfg.DbPort = 5432
	cfg.DbUser = "bhelpers_user"
	cfg.DbPassword = "123456"
	return cfg
}

func TestPostgresRedis(t *testing.T) {

	t.Skip("Run this test manually after preparing postgres and redis service.\n Don't forget to drop created databases after each run.")

	// prepare pools with postgres and redis services

	prepareCtx := initContext(t, true)

	pools1 := prepareCtx.AppWithTenancy.Pools()
	require.NotNil(t, pools1)
	require.Empty(t, pools1.Pools())
	selfPool1, err := pools1.SelfPool()
	assert.Error(t, err)
	assert.Nil(t, selfPool1)

	poolConfig1 := &TenancyPoolConfig{
		PoolName: "pool1",
	}
	poolConfig1.DbService = TenancyServiceConfig{Name: "database_service1", Type: pool.TypeDatabase, Provider: "postgres"}
	poolConfig1.DbService.DbConfig = testPostgresConfig()
	poolConfig1.DbService.DbName = "bhelpers_db"
	poolConfig1.PubsubService = TenancyServiceConfig{Name: "pubsub_service1", Type: pool.TypePubsub, Provider: pubsub_redis.Provider}
	poolConfig1.PubsubService.DbHost = "127.0.0.1"
	poolConfig1.PubsubService.DbPort = 6379
	poolConfig1.PubsubService.DbName = "0"
	p1 := preparePoolServices(t, prepareCtx, poolConfig1)

	poolConfig2 := &TenancyPoolConfig{
		PoolName: "pool2",
	}
	poolConfig2.DbService = TenancyServiceConfig{Name: "database_service2", Type: pool.TypeDatabase, Provider: "postgres"}
	poolConfig2.DbService.DbConfig = testPostgresConfig()
	poolConfig2.DbService.DbName = "bhelpers_db"
	poolConfig2.PubsubService = TenancyServiceConfig{Name: "pubsub_service2", Type: pool.TypePubsub, Provider: pubsub_redis.Provider}
	poolConfig2.PubsubService.DbHost = "127.0.0.1"
	poolConfig2.PubsubService.DbPort = 6379
	poolConfig2.PubsubService.DbName = "1"
	p2 := preparePoolServices(t, prepareCtx, poolConfig2)

	_, err = pool.ActivatePool(prepareCtx.RemotePoolController, prepareCtx.ClientOp, p1.GetID())
	require.NoError(t, err)
	_, err = pool.ActivatePool(prepareCtx.RemotePoolController, prepareCtx.ClientOp, p2.GetID())
	require.NoError(t, err)

	prepareCtx.Close()

	// prepare app with multiple pools
	multiPoolCtx := initContext(t, false)

	// add customers
	customer1, err := multiPoolCtx.LocalCustomerManager.Add(multiPoolCtx.AdminOp, "customer1", "12345678")
	require.NoError(t, err)
	require.NotNil(t, customer1)
	customer2, err := multiPoolCtx.LocalCustomerManager.Add(multiPoolCtx.AdminOp, "customer2", "12345678")
	require.NoError(t, err)
	require.NotNil(t, customer2)

	// add tenancies
	// NOTE may fail on th second run - drop all created databases befor re-run
	addedTenancy1, _ := AddTenancies(t, multiPoolCtx)

	// check if tenancy was loaded
	loadedTenancy1, err := multiPoolCtx.AppWithTenancy.Multitenancy().Tenancy(addedTenancy1.GetID())
	require.NoError(t, err)
	require.NotNil(t, loadedTenancy1)
	loadedT1, ok := loadedTenancy1.(*tenancy_manager.TenancyBase)
	require.True(t, ok)
	b2, _ := json.MarshalIndent(loadedT1.TenancyDb, "", "  ")
	t.Logf("Loaded tenancy: \n\n%s\n\n", string(b2))
	assert.Equal(t, addedTenancy1.TenancyDb, loadedT1.TenancyDb)

	// check if database tables were created
	sample1 := &InTenancySample{Field1: "hello world", Field2: 10}
	err = loadedTenancy1.Db().Create(multiPoolCtx.AdminOp, sample1)
	require.NoError(t, err)
	readSample1 := &InTenancySample{}
	found, err := loadedTenancy1.Db().FindByField(multiPoolCtx.AdminOp, "field2", 10, readSample1)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, sample1, readSample1)

	// close apps
	multiPoolCtx.Close()
	pubsub_factory.ResetSingletonInmemPubsub()

	t.Logf("Drop all created databases before re-running this test!")
}
