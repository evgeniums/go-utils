package user_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type SetPhoneEndpoint struct {
	SetUserFieldEndpoint
}

func (s *SetPhoneEndpoint) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.SetPhone")
	defer request.TraceOutMethod()

	cmd := &user_api.SetPhoneCmd{}
	err := request.ParseVerify(cmd)
	if err != nil {
		return err
	}

	err = s.users.SetPhone(request, request.GetResourceId(s.userTypeName), cmd.Phone)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func SetPhone(userTypeName string, users user.MainFieldSetters) api_server.ResourceEndpointI {
	e := &SetPhoneEndpoint{}
	return e.Init(e, userTypeName, "phone", users, user_api.SetPhone())
}
