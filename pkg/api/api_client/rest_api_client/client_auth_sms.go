package rest_api_client

import (
	"github.com/evgeniums/go-utils/pkg/api"
	"github.com/evgeniums/go-utils/pkg/op_context"
)

type ClientAuthSms struct {
	code  string
	token string
}

func (a *ClientAuthSms) HandleResponse(resp Response) {
	if resp == nil {
		return
	}
	a.token = resp.Header().Get("X-Auth-Sms-Token")
}

func (a *ClientAuthSms) SetCode(code string) {
	a.code = code
}

func (a *ClientAuthSms) MakeHeaders(ctx op_context.Context, operation api.Operation, cmd interface{}) (map[string]string, error) {

	// setup
	ctx.TraceInMethod("ClientAuthSms.MakeHeaders")
	defer ctx.TraceOutMethod()

	// no headers for empty token
	if a.token == "" {
		return nil, nil
	}

	// put code to header
	h := map[string]string{"X-Auth-Sms-Token": a.token, "X-Auth-Sms-Code": a.code}

	// clear data
	a.token = ""
	a.code = ""

	// done
	return h, nil
}

func NewClientAuthSms() *ClientAuthSms {
	c := &ClientAuthSms{}
	return c
}
