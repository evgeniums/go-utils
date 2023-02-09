package user

import "github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"

type UsersWithSession[UserType User, SessionType auth_session.Session, SessionClientType auth_session.SessionClient] interface {
	Users[UserType]
	auth_session.SessionController
}

type UsersWithSessionBase[UserType User, SessionType auth_session.Session, SessionClientType auth_session.SessionClient] struct {
	UsersBase[UserType]
	auth_session.SessionController
}

func (m *UsersWithSessionBase[UserType, SessionType, SessionClientType]) SessionManager() auth_session.SessionController {
	return m
}

type UsersWithSessionBaseConfig[UserType User] struct {
	UserController    UserController[UserType]
	SessionController auth_session.SessionController
}

func NewUsersWithSession[UserType User, SessionType auth_session.Session, SessionClientType auth_session.SessionClient](
	userBuilder func() UserType,
	sessionBuilder func() SessionType,
	sessionClientBuilder func() SessionClientType,
	config ...UsersWithSessionBaseConfig[UserType]) *UsersWithSessionBase[UserType, SessionType, SessionClientType] {

	m := &UsersWithSessionBase[UserType, SessionType, SessionClientType]{}

	if len(config) == 0 {
		m.UsersBase.Construct(LocalUserController[UserType]())
		m.SessionController = auth_session.LocalSessionController()
	} else {
		m.UsersBase.Construct(config[0].UserController)
		m.SessionController = config[0].SessionController
	}

	m.SetUserBuilder(userBuilder)
	m.SetSessionBuilder(func() auth_session.Session { return sessionBuilder() })
	m.SetSessionClientBuilder(func() auth_session.SessionClient { return sessionClientBuilder() })

	return m
}
