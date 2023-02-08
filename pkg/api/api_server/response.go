package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/api"

// Interface of response of server API.
type Response interface {
	Message() interface{}
	SetMessage(message api.Response)
	Request() Request
	SetRequest(request Request)
}

type ResponseBase struct {
	message interface{}
	request Request
}

func (r *ResponseBase) Message() interface{} {
	return r.message
}

func (r *ResponseBase) SetMessage(message api.Response) {
	if r.request.Server().IsHateoas() {
		api.InjectHateoasLinksToObject(r.request.Endpoint().Resource(), message)
	}
	r.message = message
}

func (r *ResponseBase) SetRequest(request Request) {
	r.request = request
}

func (r *ResponseBase) Request() Request {
	return r.request
}
