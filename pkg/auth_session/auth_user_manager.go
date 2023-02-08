package auth_session

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type AuthUserManager interface {
	FindAuthUser(ctx op_context.Context, login string, user interface{}, dest ...interface{}) (bool, error)
	MakeAuthUser() auth.User
	ValidateLogin(login string) error
}

type WithAuthUserManager interface {
	AuthUserManager() AuthUserManager
}

type WithUserSessionManager interface {
	WithAuthUserManager
	SessionManager() SessionController
}
