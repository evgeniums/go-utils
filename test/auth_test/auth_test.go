package auth_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_server"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/sms_provider_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_session_default"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDb(t, app, &user_session_default.User{}, &user_manager.SessionBase{}, &user_manager.SessionClientBase{})
}

func initAuthServer(t *testing.T, config ...string) (app_context.Context, *auth_server.AuthServerBase) {
	app := test_utils.InitAppContext(t, testDir, utils.OptionalArg("auth_test.json", config...))

	createDb(t, app)

	users := user_session_default.NewUsers()
	users.Init(app)

	authServer := auth_server.NewAuthServer()
	require.NoErrorf(t, authServer.Init(app, users, &sms_provider_factory.MockFactory{}), "failed to init auth server")

	return app, nil
}

func TestInitServer(t *testing.T) {
	app, _ := initAuthServer(t)
	app.Close()
}
