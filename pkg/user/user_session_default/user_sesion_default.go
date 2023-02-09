package user_session_default

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_default"
)

type User = user_default.User

type UserSession struct {
	auth_session.SessionBase
}

func NewSession() *UserSession {
	return &UserSession{}
}

type UserSessionClient struct {
	auth_session.SessionClientBase
}

func NewSessionClient() *UserSessionClient {
	return &UserSessionClient{}
}

type Users = user.UsersWithSessionBase[*user_default.User, *UserSession, *UserSessionClient]

func NewUsers(controllers ...user.UsersWithSessionBaseConfig[*User]) *Users {
	return user.NewUsersWithSession(user_default.NewUser, NewSession, NewSessionClient, controllers...)
}
