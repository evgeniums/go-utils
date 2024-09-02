package user_service

import (
	"github.com/evgeniums/go-utils/pkg/api/api_server"
	"github.com/evgeniums/go-utils/pkg/user"
	"github.com/evgeniums/go-utils/pkg/user/user_api"
)

type SetPhoneEndpoint struct {
	SetUserFieldEndpoint
}

func (s *SetPhoneEndpoint) HandleRequest(request api_server.Request) error {

	c := request.TraceInMethod("users.SetPhone")
	defer request.TraceOutMethod()

	cmd := &user.UserPhone{}
	err := request.ParseValidate(cmd)
	if err != nil {
		return err
	}

	err = Setter(s.users, request).SetPhone(request, request.GetResourceId(s.userTypeName), cmd.PHONE)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func SetPhone(userTypeName string, users user.MainFieldSetters) api_server.ResourceEndpointI {
	e := &SetPhoneEndpoint{}
	return e.Init(e, userTypeName, "phone", users, user_api.SetPhone(userTypeName))
}
