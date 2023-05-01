package auth_session

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
)

type UserValidators interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type AuthUserFinder interface {
	FindAuthUser(ctx auth.AuthContext, login string) (auth.User, error)
}

type AuthUserManager interface {
	UserValidators
	AuthUserFinder
}

type WithAuthUserManager interface {
	AuthUserManager() AuthUserManager
}

type WithUserSessionManager interface {
	WithAuthUserManager
	SessionManager() SessionController
}
