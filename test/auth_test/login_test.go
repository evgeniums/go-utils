package auth_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_server"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_session_default"
	"github.com/stretchr/testify/require"
)

func initOpTest(t *testing.T, config ...string) (app_context.Context, *user_session_default.Users, *auth_server.AuthServerBase, op_context.Context) {
	app, users, server := initAuthServer(t)

	ctx := test_utils.SimpleOpContext(app, t.Name())

	return app, users, server, ctx
}

func TestLogin(t *testing.T) {
	app, users, server, opCtx := initOpTest(t)
	defer app.Close()

	// create user1
	login1 := "user1"
	password1 := "password1"
	user1, err := users.Add(opCtx, login1, password1, user.Phone("12345678", &User{}), user.Email("user1@example.com", &User{}))
	require.NoErrorf(t, err, "failed to add user")
	require.NotNil(t, user1)

	// create user2
	login2 := "user2"
	password2 := "password2"
	user2, err := users.Add(opCtx, login2, password2, user.Phone("12345679", &User{}), user.Email("user2@example.com", &User{}))
	require.NoErrorf(t, err, "failed to add user")
	require.NotNil(t, user2)

	// prepare client
	client := test_utils.PrepareHttpClient(t, server.RestApiServer.GinEngine())

	// login without headers
	client.Login(t, login1, password1, auth.ErrorCodeUnauthorized)

	// invalid user name
	client.Login(t, "kjlkhoi43909Abc", password1, auth_login_phash.ErrorCodeLoginFailed)

	// unknown user
	client.Login(t, "someuser", password1, auth_login_phash.ErrorCodeLoginFailed)

	// good login
	client.Login(t, login1, password1)

	// re-login without headers when authorized
	client.Login(t, login2, password2, auth.ErrorCodeUnauthorized)

	// good re-login when authorized
	client.Login(t, login2, password2)

	// invalid re-login when authorized
	client.Login(t, "someuser", password1, auth_login_phash.ErrorCodeLoginFailed)
}
