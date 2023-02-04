package rest_api_client

import (
	"io"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server/rest_api_gin_server"
)

type Response interface {
	Code() int
	Header() http.Header
	Body() []byte
	Message() string

	Error() *rest_api_gin_server.ResponseError
	SetError(err *rest_api_gin_server.ResponseError)
}

type HttpResponse struct {
	Raw         *http.Response
	body        []byte
	serverError *rest_api_gin_server.ResponseError
}

func NewResponse(raw *http.Response) *HttpResponse {
	return &HttpResponse{Raw: raw, serverError: &rest_api_gin_server.ResponseError{}}
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

func (r *HttpResponse) Error() *rest_api_gin_server.ResponseError {
	return r.serverError
}

func (r *HttpResponse) SetError(err *rest_api_gin_server.ResponseError) {
	r.serverError = err
}
