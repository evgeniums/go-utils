package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type ListEndpoint[U user.User, S auth_session.Session, SC auth_session.SessionClient] struct {
	api_server.EndpointBase
	UserEndpoint[U, S, SC]
}

func List[U user.User, S auth_session.Session, SC auth_session.SessionClient](service *UserService[U, S, SC]) *ListEndpoint[U, S, SC] {
	e := &ListEndpoint[U, S, SC]{}
	e.service = service
	e.Construct(user_api.List())
	return e
}

func (e *ListEndpoint[U, S, SC]) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.List")
	defer request.TraceOutMethod()

	q := &api.DbQuery{}

	queryName := request.Endpoint().Resource().ServicePathPrototype()
	models := []interface{}{e.service.Users.MakeUser()}
	filter, err := api_server.ParseDbQuery(request, models, q, queryName)
	if err != nil {
		return c.SetError(err)
	}

	resp := &user_api.ListResponse[U]{}
	err = e.service.Users.FindUsers(request, filter, &resp.Users)
	if err != nil {
		return c.SetError(err)
	}

	if request.Server().IsHateoas() {
		api.ProcessListResourceHateousLinks(request.Endpoint().Resource(), e.service.UserTypeName, resp.Users)
	}
	request.Response().SetMessage(resp)

	return nil
}
