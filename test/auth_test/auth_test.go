package auth_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_server"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/sms_provider_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

type User struct {
	user.UserBase
}

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDb(t, app, &User{}, &user_manager.SessionBase{}, &user_manager.SessionClientBase{})
}

type UserManager struct {
	app_context.WithAppBase
	user_manager.UserManagerBase
	user_manager.SessionManagerBase
}

func NewUserManager() *UserManager {

	m := &UserManager{}
	m.MakeSession = func() user_manager.Session {
		return &user_manager.SessionBase{}
	}
	m.MakeSessionClient = func() user_manager.SessionClient {
		return &user_manager.SessionClientBase{}
	}
	return m
}

func (m *UserManager) Init(app app_context.Context) {
	m.WithAppBase.Init(app)
}

func (m *UserManager) MakeUser() auth.User {
	return &User{}
}

func (m *UserManager) SessionManager() user_manager.SessionManager {
	return m
}

func (m *UserManager) UserManager() user_manager.UserManager {
	return m
}

func (m *UserManager) ValidateLogin(login string) error {
	rules := "required,alphanum,lowercase"
	return m.App().Validator().ValidateValue(login, rules)
}

func initAuthServer(t *testing.T, config ...string) (app_context.Context, auth_server.AuthServer) {
	app := test_utils.InitAppContext(t, testDir, utils.OptionalArg("auth_test.json", config...))

	createDb(t, app)

	users := NewUserManager()
	users.Init(app)

	authServer := auth_server.NewAuthServer()
	authServer.Init(app, users, &sms_provider_factory.MockFactory{})

	return app, nil
}

func TestInitServer(t *testing.T) {
	app, _ := initAuthServer(t)
	app.Close()
}
