package rest_api_client

import (
	"io"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
)

type Response interface {
	Code() int
	Header() http.Header
	Body() []byte
	Message() string

	Error() *api.ResponseError
	SetError(err *api.ResponseError)
}

type HttpResponse struct {
	Raw         *http.Response
	body        []byte
	serverError *api.ResponseError
}

func NewResponse(raw *http.Response) *HttpResponse {
	return &HttpResponse{Raw: raw, serverError: &api.ResponseError{}}
}

func (r *HttpResponse) Code() int {
	return r.Raw.StatusCode
}

func (r *HttpResponse) Header() http.Header {
	return r.Raw.Header
}

func (r *HttpResponse) Body() []byte {
	if r.body == nil && r.Raw.Body != nil {
		r.body, _ = io.ReadAll(r.Raw.Body)
		r.Raw.Body.Close()
	}
	return r.body
}

func (r *HttpResponse) Message() string {
	return string(r.Body())
}

func (r *HttpResponse) Error() *api.ResponseError {
	return r.serverError
}

func (r *HttpResponse) SetError(err *api.ResponseError) {
	r.serverError = err
}

func IsResponseOK(resp Response, err error) bool {
	if err != nil || resp == nil {
		return false
	}
	return resp.Code() < http.StatusBadRequest
}
