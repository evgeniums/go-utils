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

func NewSession() user_manager.Session {
	return &UserSession{}
}

type UserSessionClient struct {
	user_manager.SessionClientBase
}

func NewSessionClient() user_manager.SessionClient {
	return &UserSessionClient{}
}

type Users struct {
	user.UsersWithSession[*user_default.User, *UserSession, *UserSessionClient]
}

func NewUsers(userController ...user.UserController[*User]) *Users {
	m := &Users{}
	if len(userController) == 0 {
		m.Construct(user.LocalUserController[*User]())
	} else {
		m.Construct(userController[0])
	}
	m.SetUserBuilder(user_default.NewUser)
	m.MakeSession = NewSession
	m.MakeSessionClient = NewSessionClient
	return m
}
