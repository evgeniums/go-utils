package user

import (
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type UsersWithSession[UserType User, SessionType user_manager.Session, SessionClientType user_manager.SessionClient] struct {
	UsersBase[UserType]
	user_manager.SessionController
}

func (m *UsersWithSession[UserType, SessionType, SessionClientType]) SessionManager() user_manager.SessionController {
	return m
}

type UsersWithSessionConfig[UserType User] struct {
	UserController    UserController[UserType]
	SessionController user_manager.SessionController
}

func NewUsersWithSession[UserType User, SessionType user_manager.Session, SessionClientType user_manager.SessionClient](
	userBuilder func() UserType,
	sessionBuilder func() SessionType,
	sessionClientBuilder func() SessionClientType,
	config ...UsersWithSessionConfig[UserType]) *UsersWithSession[UserType, SessionType, SessionClientType] {

	m := &UsersWithSession[UserType, SessionType, SessionClientType]{}

	if len(config) == 0 {
		m.UsersBase.Construct(LocalUserController[UserType]())
		m.SessionController = user_manager.LocalSessionController()
	} else {
		m.UsersBase.Construct(config[0].UserController)
		m.SessionController = config[0].SessionController
	}

	m.SetUserBuilder(userBuilder)
	m.SetSessionBuilder(func() user_manager.Session { return sessionBuilder() })
	m.SetSessionClientBuilder(func() user_manager.SessionClient { return sessionClientBuilder() })

	return m
}
