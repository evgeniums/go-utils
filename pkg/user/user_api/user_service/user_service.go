package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type UserEndpoint[U user.User, S user_manager.Session, SC user_manager.SessionClient] struct {
	service *UserService[U, S, SC]
}

type UserService[U user.User, S user_manager.Session, SC user_manager.SessionClient] struct {
	api_server.ServiceBase
	Users *user.UsersWithSession[U, S, SC]
}

func NewUserService[U user.User, S user_manager.Session, SC user_manager.SessionClient](userController *user.UsersWithSession[U, S, SC],
	UName ...string) *UserService[U, S, SC] {

	serviceName, users, user := user_api.PrepareResources(UName...)
	s := &UserService[U, S, SC]{}
	s.Init(serviceName)
	s.Users = userController
	s.AddChild(users)

	users.AddOperation(Find(s))

	user.AddOperation(Update(s))

	return s
}
