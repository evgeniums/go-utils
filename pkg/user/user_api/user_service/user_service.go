package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type UserEndpoint[U user.User, S auth_session.Session, SC auth_session.SessionClient] struct {
	service *UserService[U, S, SC]
}

type UserService[U user.User, S auth_session.Session, SC auth_session.SessionClient] struct {
	api_server.ServiceBase
	Users        *user.UsersWithSession[U, S, SC]
	UserTypeName string
}

func NewUserService[U user.User, S auth_session.Session, SC auth_session.SessionClient](userController *user.UsersWithSession[U, S, SC],
	UserTypeName ...string) *UserService[U, S, SC] {

	userType, serviceName, users, user := user_api.PrepareResources(UserTypeName...)
	s := &UserService[U, S, SC]{}
	s.Init(serviceName)
	s.Users = userController
	s.AddChild(users)
	s.UserTypeName = userType

	users.AddOperation(List(s))

	user.AddOperation(Update(s))

	return s
}
