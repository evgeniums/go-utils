package user_manager

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
)

type UserManager interface {
	UserController

	MakeAuthUser() auth.User
	ValidateLogin(login string) error
}

type WithUserManager interface {
	UserManager() UserManager
}

type WithUserSessionManager interface {
	WithUserManager
	SessionManager() SessionController
}
