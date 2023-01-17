package api_server

// Interface of response of server API.
type Response interface {
	Message() interface{}
}

type ResponseBase struct {
	message interface{}
}

func (r *ResponseBase) Message() interface{} {
	return r.message
}
