package auth_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_server"
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

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDb(t, app, &User{}, &user_session_default.UserSession{}, &user_session_default.UserSessionClient{}, &sms.SmsMessage{})
}

func initAuthServer(t *testing.T, config ...string) (app_context.Context, *user_session_default.Users, *auth_server.AuthServerBase) {
	app := test_utils.InitAppContext(t, testDir, utils.OptionalArg("auth_test.jsonc", config...))

	createDb(t, app)

	users := user_session_default.NewUsers()
	users.Init(app)

	server := auth_server.NewAuthServer()
	require.NoErrorf(t, server.Init(app, users, &sms_provider_factory.MockFactory{}), "failed to init auth server")

	return app, users, server
}

func TestInitServer(t *testing.T) {
	app, _, _ := initAuthServer(t)
	app.Close()
}
