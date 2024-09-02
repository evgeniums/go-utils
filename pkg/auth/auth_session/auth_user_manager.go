package auth_session

import (
	"github.com/evgeniums/go-utils/pkg/auth"
	"github.com/evgeniums/go-utils/pkg/op_context"
)

type UserValidators interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type AuthUserFinder interface {
	FindAuthUser(ctx op_context.Context, login string) (auth.User, error)
}

type AuthUserManager interface {
	UserValidators
	AuthUserFinder
}

type WithAuthUserManager interface {
	AuthUserManager() AuthUserManager
}

type WithSessionManager interface {
	SessionManager() SessionController
}

type WithUserSessionManager interface {
	WithAuthUserManager
	WithSessionManager
}
