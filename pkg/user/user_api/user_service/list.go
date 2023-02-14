package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ListEndpoint[U user.User] struct {
	api_server.EndpointBase
	UserEndpoint[U]
}

func (e *ListEndpoint[U]) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.List")
	defer request.TraceOutMethod()

	u := Users(e.service, request)

	cmd := &api.DbQuery{}
	queryName := request.Endpoint().Resource().ServicePathPrototype()
	filter, err := api_server.ParseDbQuery(request, utils.List(u.MakeUser()), cmd, queryName)
	if err != nil {
		return c.SetError(err)
	}

	resp := &user_api.ListResponse[U]{}
	users := make([]U, 0)
	resp.Users = &users
	resp.Count, err = u.FindUsers(request, filter, resp.Users)
	if err != nil {
		return c.SetError(err)
	}

	if request.Server().IsHateoas() {
		api.ProcessListResourceHateousLinks(request.Endpoint().Resource(), e.service.UserTypeName, *resp.Users)
	}
	request.Response().SetMessage(resp)

	return nil
}

func List[U user.User](service *UserService[U]) *ListEndpoint[U] {
	e := &ListEndpoint[U]{}
	e.service = service
	e.Construct(user_api.List())
	return e
}
