package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
	"github.com/evgeniums/go-backend-helpers/pkg/user_manager"
)

type FindEndpoint[U user.User, S user_manager.Session, SC user_manager.SessionClient] struct {
	api_server.EndpointBase
	UserEndpoint[U, S, SC]
}

func Find[U user.User, S user_manager.Session, SC user_manager.SessionClient](service *UserService[U, S, SC]) *FindEndpoint[U, S, SC] {
	e := &FindEndpoint[U, S, SC]{}
	e.service = service
	e.Construct(user_api.Find())
	return e
}

func (e *FindEndpoint[U, S, SC]) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.Find")
	defer request.TraceOutMethod()

	q := &api.DbQuery{}

	queryName := request.Endpoint().Resource().ServicePathPrototype()
	models := []interface{}{e.service.Users.MakeUser()}
	filter, err := api_server.ParseDbQuery(request, models, q, queryName)
	if err != nil {
		return c.SetError(err)
	}

	var users []U
	err = e.service.Users.FindUsers(request, filter, &users)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}
