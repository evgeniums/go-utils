package pool_api_test

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_service"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func dbModels() []interface{} {
	return append([]interface{}{}, admin.DbModels(), pool.DbModels())
}

type testContext struct {
	*api_test.TestContext

	LocalPoolController  pool.PoolController
	RemotePoolController *pool_client.PoolClient
}

func initTest(t *testing.T) *testContext {

	ctx := &testContext{}

	ctx.TestContext = api_test.InitTest(t, "pool", testDir, dbModels())
	ctx.RemotePoolController = pool_client.NewPoolClient(ctx.RestApiClient)
	p := pool.NewPoolController(&crud.DbCRUD{})
	ctx.LocalPoolController = p

	poolService := pool_service.NewPoolService(ctx.LocalPoolController)
	api_server.AddServiceToServer(ctx.Server.ApiServer(), poolService)

	return ctx
}

func TestInit(t *testing.T) {
	ctx := initTest(t)
	ctx.Close()
}

func addPool(t *testing.T, ctx *testContext) pool.Pool {

	p1Sample := &pool.PoolBaseData{}
	p1Sample.SetName("pool1")
	p1Sample.SetLongName("pool1 long name")
	p1Sample.SetDescription("pool description")

	p1 := pool.NewPool()
	p1.SetName(p1Sample.Name())
	p1.SetDescription(p1Sample.Description())
	p1.SetLongName(p1Sample.LongName())
	addedPool1, err := ctx.RemotePoolController.AddPool(ctx.ClientOp, p1)
	require.NoError(t, err)
	require.NotNil(t, addedPool1)
	assert.Equal(t, p1Sample.Name(), addedPool1.Name())
	assert.Equal(t, p1Sample.LongName(), addedPool1.LongName())
	assert.Equal(t, p1Sample.Description(), addedPool1.Description())
	assert.NotEmpty(t, addedPool1.GetID())

	dbPool1, err := ctx.LocalPoolController.FindPool(ctx.AdminOp, p1Sample.Name(), true)
	require.NoError(t, err)
	require.NotNil(t, dbPool1)

	b1, _ := json.Marshal(addedPool1)
	b2, _ := json.Marshal(dbPool1)
	assert.Equal(t, string(b1), string(b2))

	return addedPool1
}

func TestAddPool(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	addPool(t, ctx)
}

func addService(t *testing.T, ctx *testContext) pool.PoolService {
	p1Sample := &pool.PoolServiceBaseEssentials{}
	p1Sample.SetName("service1")
	p1Sample.SetLongName("service1 long name")
	p1Sample.SetDescription("service description")
	p1Sample.SetType("database")
	p1Sample.SetRefId("reference id")

	p1Sample.ServiceConfigBase.PUBLIC_HOST = "pubhost"
	p1Sample.ServiceConfigBase.PUBLIC_PORT = 1122
	p1Sample.ServiceConfigBase.PUBLIC_URL = "puburl"
	p1Sample.ServiceConfigBase.PRIVATE_HOST = "privhost"
	p1Sample.ServiceConfigBase.PRIVATE_PORT = 3344
	p1Sample.ServiceConfigBase.PRIVATE_URL = "privurl"
	p1Sample.ServiceConfigBase.PARAMETER1 = "param1"
	p1Sample.ServiceConfigBase.PARAMETER2 = "param2"
	p1Sample.ServiceConfigBase.PARAMETER3 = "param3"

	p1 := pool.NewService()
	p1.SetName(p1Sample.Name())
	p1.SetLongName(p1Sample.LongName())
	p1.SetDescription(p1Sample.Description())
	p1.SetType(p1Sample.Type())
	p1.SetRefId(p1Sample.RefId())
	p1.PoolServiceBaseEssentials.ServiceConfigBase = p1Sample.ServiceConfigBase
	p1.SECRET1 = "secret1"
	p1.SECRET2 = "secret2"

	addedService1, err := ctx.RemotePoolController.AddService(ctx.ClientOp, p1)
	require.NoError(t, err)
	require.NotNil(t, addedService1)
	assert.NotEmpty(t, addedService1.GetID())
	addedB1, ok := addedService1.(*pool.PoolServiceBase)
	require.True(t, ok)
	assert.Equal(t, p1.PoolServiceBaseEssentials, addedB1.PoolServiceBaseEssentials)
	assert.Equal(t, p1.Secret1(), addedService1.Secret1())
	assert.Equal(t, p1.Secret2(), addedService1.Secret2())

	dbService1, err := ctx.LocalPoolController.FindService(ctx.AdminOp, p1Sample.Name(), true)
	require.NoError(t, err)
	require.NotNil(t, dbService1)

	b1, _ := json.Marshal(addedService1)
	b2, _ := json.Marshal(dbService1)
	assert.Equal(t, string(b1), string(b2))

	return addedService1
}

func TestAddService(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	addService(t, ctx)
}

func TestGetBindings(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	ctx.AdminOp.Db().EnableDebug(true)

	service := addService(t, ctx)
	p := addPool(t, ctx)
	role := "main_webservice"

	err := ctx.LocalPoolController.AddServiceToPool(ctx.AdminOp, p.GetID(), service.GetID(), role)
	require.NoError(t, err)

	bindings, err := ctx.LocalPoolController.GetPoolBindings(ctx.AdminOp, p.GetID())
	require.NoError(t, err)
	require.Equal(t, 1, len(bindings))

	binding := bindings[0]
	assert.Equal(t, p.Name(), binding.PoolName)
	assert.Equal(t, p.GetID(), binding.PoolId)
	assert.Equal(t, service.Name(), binding.ServiceName)
	assert.Equal(t, service.GetID(), binding.ServiceId)
	assert.Equal(t, role, binding.Role())

	serviceB := service.(*pool.PoolServiceBase)
	assert.Equal(t, serviceB.PoolServiceBaseData, binding.PoolServiceBaseData)

	b1, _ := json.MarshalIndent(bindings[0], "", "  ")
	t.Logf("Pool services: \n\n%s\n\n", string(b1))

	bindings, err = ctx.LocalPoolController.GetServiceBindings(ctx.AdminOp, service.GetID())
	require.NoError(t, err)
	require.Equal(t, 1, len(bindings))
}
