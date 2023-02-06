package api_server

import "github.com/evgeniums/go-backend-helpers/pkg/generic_error"

// Interface of response of server API.
type Response interface {
	Message() interface{}
	SetMessage(message interface{})
}

type ResponseBase struct {
	message interface{}
}

func (r *ResponseBase) Message() interface{} {
	return r.message
}

func (r *ResponseBase) SetMessage(message interface{}) {
	r.message = message
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func ResponseGenericError(responseError *ResponseError) generic_error.Error {
	if responseError == nil {
		return nil
	}
	e := generic_error.New(responseError.Code, responseError.Message)
	e.SetDetails(responseError.Details)
	return e
}
