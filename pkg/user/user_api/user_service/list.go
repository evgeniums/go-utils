package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type ListEndpoint[U user.User] struct {
	api_server.EndpointBase
	UserEndpoint[U]
}

func (e *ListEndpoint[U]) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.List")
	defer request.TraceOutMethod()

	u := Users(e.service, request)

	queryName := request.Endpoint().Resource().ServicePathPrototype()
	filter, err := api_server.ParseDbQuery(request, u.MakeUser(), queryName)
	if err != nil {
		return c.SetError(err)
	}

	resp := &user_api.ListResponse[U]{}
	resp.Items, resp.Count, err = u.FindUsers(request, filter)
	if err != nil {
		return c.SetError(err)
	}

	/*
		// TODO process hateous links
		if request.Server().IsHateoas() {
			api.ProcessListResourceHateousLinks(request.Endpoint().Resource(), e.service.UserTypeName, resp.Items)
		}
	*/
	request.Response().SetMessage(resp)

	return nil
}

func List[U user.User](service *UserService[U]) *ListEndpoint[U] {
	e := &ListEndpoint[U]{}
	e.service = service
	e.Construct(user_api.List())
	return e
}
