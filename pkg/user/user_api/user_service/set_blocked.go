package user_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/user"
	"github.com/evgeniums/go-utils/pkg/user/user_api"
)

type SetBlockedEndpoint struct {
	SetUserFieldEndpoint
}

func (s *SetBlockedEndpoint) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.SetBlocked")
	defer request.TraceOutMethod()

	cmd := &user.UserBlocked{}
	err := request.ParseValidate(cmd)
	if err != nil {
		return err
	}

	err = Setter(s.users, request).SetBlocked(request, request.GetResourceId(s.userTypeName), cmd.BLOCKED)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func SetBlocked(userTypeName string, users user.MainFieldSetters) api_server.ResourceEndpointI {
	e := &SetBlockedEndpoint{}
	return e.Init(e, userTypeName, "blocked", users, user_api.SetBlocked(userTypeName))
}
