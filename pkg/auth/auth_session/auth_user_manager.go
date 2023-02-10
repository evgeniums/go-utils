package auth_session

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type UserValidators interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type AuthUserManager interface {
	UserValidators
	FindAuthUser(ctx op_context.Context, login string, user auth.User, dest ...interface{}) (bool, error)
	MakeAuthUser() auth.User
}

type WithAuthUserManager interface {
	AuthUserManager() AuthUserManager
}

type WithUserSessionManager interface {
	WithAuthUserManager
	SessionManager() SessionController
}
