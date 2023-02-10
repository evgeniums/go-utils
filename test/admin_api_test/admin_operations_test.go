package admin_api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/admin/admin_api/admin_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/bare_bones_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var targetAdminLogin = "target_admin"
var targetAdminPassword = "target_admin_password"
var targetAdminPhone = "888999000"
var targetAdminEmail = "target@example.com"

type TestContext struct {
	ClientApp          app_context.Context
	ServerApp          app_context.Context
	AdminClient        *admin_api_client.AdminClient
	ClientOp           op_context.Context
	AdminOp            op_context.Context
	Server             bare_bones_server.Server
	LocalAdminManager  *admin.Manager
	RemoteAdminManager *admin.Manager

	TargetUser *Admin
}

func (t *TestContext) Close() {
	t.ClientOp.Close()
	t.AdminOp.Close()
	t.ClientApp.Close()
	t.ServerApp.Close()
}

func (t *TestContext) Reset() {
	t.ClientOp.Reset()
	t.AdminOp.Reset()
}

func initTest(t *testing.T) *TestContext {

	ctx := &TestContext{}

	ctx.ServerApp, ctx.LocalAdminManager, ctx.Server = initServer(t)
	ctx.ClientApp, ctx.AdminClient = initClient(t, test_utils.BBGinEngine(t, ctx.Server))

	ctx.ClientOp = test_utils.SimpleOpContext(ctx.ClientApp, t.Name())
	ctx.AdminOp = test_utils.SimpleOpContext(ctx.ServerApp, t.Name())

	ctx.RemoteAdminManager = admin.NewManager(admin.AdminControllers{UserController: ctx.AdminClient})

	// add superadmin for remote admin manager login
	superadmin := "superadmin"
	superpassword := "superpassword"
	user1, err := ctx.LocalAdminManager.Add(ctx.AdminOp, superadmin, superpassword)
	require.NoErrorf(t, err, "failed to add superadmin")
	require.NotNil(t, user1)

	// login with client
	restApiClient, ok := ctx.AdminClient.Client().Transport().(rest_api_client.RestApiClient)
	require.True(t, ok)
	resp, err := restApiClient.Login(ctx.ClientOp, superadmin, superpassword)
	require.NoErrorf(t, err, "failed to login superadmin")
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.Code())

	// add target admin for various user operations
	ctx.TargetUser, err = ctx.LocalAdminManager.Add(ctx.AdminOp, targetAdminLogin, targetAdminPassword, user.Phone(targetAdminPhone, &Admin{}), user.Email(targetAdminEmail, &Admin{}))
	require.NoErrorf(t, err, "failed to add target admin")
	require.NotNil(t, ctx.TargetUser)

	return ctx
}

func TestAdd(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	login1 := "admin1"
	password1 := "admin_password1"
	phone1 := "999000111"
	addedAdmin1, err := ctx.RemoteAdminManager.AddAdmin(ctx.ClientOp, login1, password1, phone1)
	require.NoError(t, err)
	require.NotNil(t, addedAdmin1)
	assert.Equal(t, login1, addedAdmin1.Login())
	assert.Equal(t, phone1, addedAdmin1.Phone())

	dbAdmin1, err := ctx.LocalAdminManager.FindByLogin(ctx.AdminOp, login1)
	require.NoError(t, err)
	require.NotNil(t, dbAdmin1)

	b1, _ := json.Marshal(addedAdmin1)
	b2, _ := json.Marshal(dbAdmin1)
	assert.Equal(t, b1, b2)

	login2 := "admin2"
	password2 := "admin_password2"
	phone2 := "999000222"
	email2 := "admin2@example.com"
	addedAdmin2, err := ctx.RemoteAdminManager.Add(ctx.ClientOp, login2, password2, user.Phone(phone2, &Admin{}), user.Email(email2, &Admin{}))
	require.NoError(t, err)
	require.NotNil(t, addedAdmin2)
	assert.Equal(t, login2, addedAdmin2.Login())
	assert.Equal(t, phone2, addedAdmin2.Phone())
	assert.Equal(t, email2, addedAdmin2.Email())

	dbAdmin2, err := ctx.LocalAdminManager.FindByLogin(ctx.AdminOp, login2)
	require.NoError(t, err)
	require.NotNil(t, dbAdmin2)

	c1, _ := json.Marshal(addedAdmin2)
	c2, _ := json.Marshal(dbAdmin2)
	assert.Equal(t, c1, c2)

	restClient1 := test_utils.PrepareHttpClient(t, test_utils.BBGinEngine(t, ctx.Server))
	restClient1.Login(login1, password1)
	restClient2 := test_utils.PrepareHttpClient(t, test_utils.BBGinEngine(t, ctx.Server))
	restClient2.Login(login2, password2)
}

func TestSetPhone(t *testing.T) {
	ctx := initTest(t)
	defer ctx.Close()

	newPhone := "888999111"
	err := ctx.RemoteAdminManager.SetPhone(ctx.ClientOp, targetAdminLogin, newPhone)
	assert.Error(t, err)
	gErr, ok := err.(generic_error.Error)
	require.True(t, ok)
	assert.Equal(t, generic_error.ErrorCodeNotFound, gErr.Code())
	ctx.Reset()

	dbAdmin1, err := ctx.LocalAdminManager.FindByLogin(ctx.AdminOp, targetAdminLogin)
	require.NoError(t, err)
	require.NotNil(t, dbAdmin1)
	assert.Equal(t, targetAdminPhone, dbAdmin1.Phone())
	ctx.Reset()

	assert.NoError(t, ctx.RemoteAdminManager.SetPhone(ctx.ClientOp, ctx.TargetUser.GetID(), newPhone))
	dbAdmin2, err := ctx.LocalAdminManager.FindByLogin(ctx.AdminOp, targetAdminLogin)
	require.NoError(t, err)
	require.NotNil(t, dbAdmin2)
	assert.Equal(t, newPhone, dbAdmin2.Phone())
}
