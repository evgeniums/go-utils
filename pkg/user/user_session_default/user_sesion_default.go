package user_session_default

import (
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_default"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type User = user_default.User

type UserSession struct {
	user_manager.SessionBase
}

func NewSession() *UserSession {
	return &UserSession{}
}

type UserSessionClient struct {
	user_manager.SessionClientBase
}

func NewSessionClient() *UserSessionClient {
	return &UserSessionClient{}
}

type Users = user.UsersWithSession[*user_default.User, *UserSession, *UserSessionClient]

func NewUsers(controllers ...user.UsersWithSessionConfig[*User]) *Users {
	return user.NewUsersWithSession(user_default.NewUser, NewSession, NewSessionClient, controllers...)
}
