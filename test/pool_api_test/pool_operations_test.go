package pool_api_test

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/app_with_pools"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_service"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func dbModels() []interface{} {
	return utils.ConcatSlices([]interface{}{}, admin.DbModels(), pool.DbModels())
}

type testContext struct {
	*api_test.TestContext

	LocalPoolController  pool.PoolController
	RemotePoolController *pool_client.PoolClient
}

func initTest(t *testing.T) *testContext {

	var appWithPools *app_with_pools.AppWithPoolsBase

	buildApp := func(t *testing.T, buildConfig *app_context.BuildConfig) app_context.Context {
		appWithPools = app_with_pools.New(buildConfig)
		return appWithPools
	}
	initApp := func(t *testing.T, app app_context.Context, configFile string, args []string, configType ...string) error {
		a, ok := app.(*app_with_pools.AppWithPoolsBase)
		require.True(t, ok)
		opCtx, err := a.InitWithArgs(configFile, args, configType...)
		if opCtx != nil {
			opCtx.Close()
		}
		return err
	}
	test_utils.SetAppHandlers(buildApp, initApp)

	ctx := &testContext{}
	ctx.TestContext = api_test.InitTest(t, "pool", testDir, dbModels())
	ctx.RemotePoolController = pool_client.NewPoolClient(ctx.RestApiClient)
	ctx.LocalPoolController = appWithPools.Pools().PoolController()

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

	remotePool1, err := ctx.RemotePoolController.FindPool(ctx.AdminOp, p1Sample.Name(), true)
	require.NoError(t, err)
	require.NotNil(t, dbPool1)

	b3, _ := json.Marshal(remotePool1)
	assert.Equal(t, string(b1), string(b3))

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
	p1Sample.SetTypeName("database")
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
	p1.SetTypeName(p1Sample.TypeName())
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

	remoteService1, err := ctx.RemotePoolController.FindService(ctx.AdminOp, p1Sample.Name(), true)
	require.NoError(t, err)
	require.NotNil(t, remoteService1)

	b3, _ := json.Marshal(remoteService1)
	assert.Equal(t, string(b1), string(b3))

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

	checkList := func(bindings []*pool.PoolServiceBinding) {
		require.Equal(t, 1, len(bindings))
		binding := bindings[0]
		assert.Equal(t, p.Name(), binding.PoolName)
		assert.Equal(t, p.GetID(), binding.PoolId)
		assert.Equal(t, service.Name(), binding.ServiceName)
		assert.Equal(t, service.GetID(), binding.ServiceId)
		assert.Equal(t, role, binding.Role())
		serviceB := service.(*pool.PoolServiceBase)
		assert.Equal(t, serviceB.PoolServiceBaseData, binding.PoolServiceBaseData)
	}

	bindings, err := ctx.LocalPoolController.GetPoolBindings(ctx.AdminOp, p.GetID())
	require.NoError(t, err)
	checkList(bindings)

	b1, _ := json.MarshalIndent(bindings[0], "", "  ")
	t.Logf("Pool services: \n\n%s\n\n", string(b1))

	bindings, err = ctx.LocalPoolController.GetServiceBindings(ctx.AdminOp, service.GetID())
	require.NoError(t, err)
	checkList(bindings)
}

func TestUpdatePool(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	p := addPool(t, ctx)

	// update and check pool
	fields := db.Fields{"name": "updated name", "long_name": "updated long_name", "description": "updated description", "active": false}
	updatedP, err := ctx.RemotePoolController.UpdatePool(ctx.ClientOp, p.GetID(), fields)
	require.NoError(t, err)
	assert.Equal(t, "updated name", updatedP.Name())
	assert.Equal(t, "updated long_name", updatedP.LongName())
	assert.Equal(t, "updated description", updatedP.Description())
	assert.False(t, updatedP.IsActive())

	// find and check pool
	remotePool1, err := ctx.RemotePoolController.FindPool(ctx.AdminOp, p.GetID())
	require.NoError(t, err)
	require.NotNil(t, remotePool1)
	assert.Equal(t, "updated name", remotePool1.Name())
	assert.Equal(t, "updated long_name", remotePool1.LongName())
	assert.Equal(t, "updated description", remotePool1.Description())
	assert.False(t, remotePool1.IsActive())

	// unknown pool
	delete(fields, "name")
	_, err = ctx.RemotePoolController.UpdatePool(ctx.ClientOp, "someid", fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, pool.ErrorCodePoolNotFound, "Pool not found.")

	// unknown field
	fields = db.Fields{"unknown_field": "try me"}
	_, err = ctx.RemotePoolController.UpdatePool(ctx.ClientOp, p.GetID(), fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, generic_error.ErrorCodeFormat, "Invalid fields for update.")

	// duplicate name
	fields = db.Fields{"name": "updated name"}
	_, err = ctx.RemotePoolController.UpdatePool(ctx.ClientOp, p.GetID(), fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, pool.ErrorCodePoolNameConflict, "Pool with such name already exists, choose another name.")
}

func TestUpdateService(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	s := addService(t, ctx)

	// fiil fields
	fields := db.Fields{"name": "updated name", "long_name": "updated long_name", "description": "updated description", "active": false}
	fields["type_name"] = "new type"
	fields["secret1"] = "new secret 1"
	fields["secret2"] = "new secret 2"
	fields["provider"] = "new provider"
	fields["public_host"] = "new public host"
	fields["public_port"] = 1010
	fields["public_url"] = "new public url"
	fields["private_host"] = "new private host"
	fields["private_port"] = 2020
	fields["private_url"] = "new private url"
	fields["user"] = "new user"
	fields["parameter1"] = "new parameter 1"
	fields["parameter2"] = "new parameter 2"
	fields["parameter3"] = "new parameter 3"
	fields["parameter1_name"] = "new name of parameter 1"
	fields["parameter2_name"] = "new name of parameter 2"
	fields["parameter3_name"] = "new name of parameter 3"

	// update and check service
	updatedS, err := ctx.RemotePoolController.UpdateService(ctx.ClientOp, s.GetID(), fields)
	require.NoError(t, err)
	assert.Equal(t, "updated name", updatedS.Name())
	assert.Equal(t, "updated long_name", updatedS.LongName())
	assert.Equal(t, "updated description", updatedS.Description())
	assert.Equal(t, "new type", updatedS.TypeName())
	assert.Equal(t, "new secret 1", updatedS.Secret1())
	assert.Equal(t, "new secret 2", updatedS.Secret2())
	assert.Equal(t, "new provider", updatedS.Provider())
	assert.Equal(t, "new public host", updatedS.PublicHost())
	assert.Equal(t, uint16(1010), updatedS.PublicPort())
	assert.Equal(t, "new public url", updatedS.PublicUrl())
	assert.Equal(t, "new private host", updatedS.PrivateHost())
	assert.Equal(t, uint16(2020), updatedS.PrivatePort())
	assert.Equal(t, "new private url", updatedS.PrivateUrl())
	assert.Equal(t, "new user", updatedS.User())
	assert.Equal(t, "new parameter 1", updatedS.Parameter1())
	assert.Equal(t, "new parameter 2", updatedS.Parameter2())
	assert.Equal(t, "new parameter 3", updatedS.Parameter3())
	assert.Equal(t, "new name of parameter 1", updatedS.Parameter1Name())
	assert.Equal(t, "new name of parameter 2", updatedS.Parameter2Name())
	assert.Equal(t, "new name of parameter 3", updatedS.Parameter3Name())
	assert.False(t, updatedS.IsActive())

	// find and check service
	remoteService1, err := ctx.RemotePoolController.FindService(ctx.AdminOp, s.GetID())
	require.NoError(t, err)
	require.NotNil(t, remoteService1)
	assert.Equal(t, "updated name", remoteService1.Name())
	assert.Equal(t, "updated long_name", remoteService1.LongName())
	assert.Equal(t, "updated description", remoteService1.Description())
	assert.Equal(t, "new type", remoteService1.TypeName())
	assert.Equal(t, "new secret 1", remoteService1.Secret1())
	assert.Equal(t, "new secret 2", remoteService1.Secret2())
	assert.Equal(t, "new provider", remoteService1.Provider())
	assert.Equal(t, "new public host", remoteService1.PublicHost())
	assert.Equal(t, uint16(1010), remoteService1.PublicPort())
	assert.Equal(t, "new public url", remoteService1.PublicUrl())
	assert.Equal(t, "new private host", remoteService1.PrivateHost())
	assert.Equal(t, uint16(2020), remoteService1.PrivatePort())
	assert.Equal(t, "new private url", remoteService1.PrivateUrl())
	assert.Equal(t, "new user", remoteService1.User())
	assert.Equal(t, "new parameter 1", remoteService1.Parameter1())
	assert.Equal(t, "new parameter 2", remoteService1.Parameter2())
	assert.Equal(t, "new parameter 3", remoteService1.Parameter3())
	assert.Equal(t, "new name of parameter 1", remoteService1.Parameter1Name())
	assert.Equal(t, "new name of parameter 2", remoteService1.Parameter2Name())
	assert.Equal(t, "new name of parameter 3", remoteService1.Parameter3Name())
	assert.False(t, remoteService1.IsActive())

	// unknown service
	delete(fields, "name")
	_, err = ctx.RemotePoolController.UpdateService(ctx.ClientOp, "someid", fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, pool.ErrorCodeServiceNotFound, "Service not found.")

	// unknown field
	fields = db.Fields{"unknown_field": "try me"}
	_, err = ctx.RemotePoolController.UpdateService(ctx.ClientOp, s.GetID(), fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, generic_error.ErrorCodeFormat, "Invalid fields for update.")

	// duplicate name
	fields = db.Fields{"name": "updated name"}
	_, err = ctx.RemotePoolController.UpdateService(ctx.ClientOp, s.GetID(), fields)
	require.Error(t, err)
	test_utils.CheckGenericError(t, err, pool.ErrorCodeServiceNameConflict, "Service with such name already exists, choose another name.")
}
