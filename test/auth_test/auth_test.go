package auth_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api/bare_bones_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/sms_provider_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_default"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_session_default"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

type User = user_default.User

func dbModels() []interface{} {
	return append([]interface{}{}, &User{}, &user_session_default.UserSession{}, &user_session_default.UserSessionClient{}, &sms.SmsMessage{})
}

func initServer(t *testing.T, config ...string) (app_context.Context, *user_session_default.Users, bare_bones_server.Server) {
	app := test_utils.InitAppContext(t, testDir, dbModels(), utils.OptionalArg("auth_test.jsonc", config...))

	users := user_session_default.NewUsers()
	users.Init(app.Validator())

	tenancyManager := &tenancy_manager.TenancyManager{}

	server := bare_bones_server.New(users, bare_bones_server.Config{SmsProviders: &sms_provider_factory.MockFactory{}})
	require.NoErrorf(t, server.Init(app, tenancyManager), "failed to init auth server")

	return app, users, server
}

func TestInitServer(t *testing.T) {
	app, _, _ := initServer(t)
	app.Close()
}
