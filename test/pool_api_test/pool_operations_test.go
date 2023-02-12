package pool_api_test

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_client"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/pool_api/pool_service"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/test/api_test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDbModels(t, app, admin.DbModels())
	test_utils.CreateDbModels(t, app, pool.DbModels())
}

type testContext struct {
	*api_test.TestContext

	LocalPoolController  pool.PoolController
	RemotePoolController *pool_client.PoolClient
}

func initTest(t *testing.T) *testContext {

	ctx := &testContext{}

	ctx.TestContext = api_test.InitTest(t, "pool", testDir, createDb)
	ctx.RemotePoolController = pool_client.NewPoolClient(ctx.RestApiClient)
	p := pool.NewPoolController(&crud.DbCRUD{})
	ctx.LocalPoolController = p

	poolService := pool_service.NewPoolService(ctx.LocalPoolController)
	api_server.AddServiceToServer(ctx.Server.ApiServer(), poolService)

	return ctx
}

func TestInit(t *testing.T) {
	initTest(t)
}

func TestAddPool(t *testing.T) {
	ctx := initTest(t)

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
}
