package rest_api_gin_server

import "github.com/evgeniums/go-backend-helpers/pkg/api_server"

type Response struct {
	api_server.ResponseBase

	httpCode int
	request  *Request
}

type ResponseError = api_server.ResponseError
