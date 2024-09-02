package admin_api_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-utils/pkg/admin"
	"github.com/evgeniums/go-utils/pkg/admin/admin_api/admin_api_service"
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/api/bare_bones_server"
	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-utils/pkg/sms"
	"github.com/evgeniums/go-utils/pkg/sms/sms_provider_factory"
	"github.com/evgeniums/go-utils/pkg/test_utils"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

type Admin = admin.Admin

func dbModels() []interface{} {
	return append(admin.DbModels(), &sms.SmsMessage{})
}

func initServer(t *testing.T, config ...string) (app_context.Context, *admin.Manager, bare_bones_server.Server) {
	app := test_utils.InitAppContext(t, testDir, dbModels(), utils.OptionalArg("admin_api_server.jsonc", config...))

	admins := admin.NewManager()
	admins.Init(app.Validator())

	tenancyManager := &tenancy_manager.TenancyManager{}

	server := bare_bones_server.New(admins, bare_bones_server.Config{SmsProviders: &sms_provider_factory.MockFactory{}})
	require.NoErrorf(t, server.Init(app, tenancyManager), "failed to init server")

	adminService := admin_api_service.NewAdminService(admins)
	api_server.AddServiceToServer(server.ApiServer(), adminService)

	return app, admins, server
}

func TestInitServer(t *testing.T) {
	app, _, _ := initServer(t)
	app.Close()
}
