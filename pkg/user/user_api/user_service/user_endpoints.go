package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type UpdateEndpoint[U user.User, S user_manager.Session, SC user_manager.SessionClient] struct {
	api_server.EndpointBase
	UserEndpoint[U, S, SC]
}

func (e *UpdateEndpoint[U, S, SC]) HandleRequest(request api_server.Request) error {
	return nil
}

func Update[U user.User, S user_manager.Session, SC user_manager.SessionClient](service *UserService[U, S, SC]) *UpdateEndpoint[U, S, SC] {
	e := &UpdateEndpoint[U, S, SC]{}
	e.service = service
	e.Construct(user_api.Update())
	return e
}
