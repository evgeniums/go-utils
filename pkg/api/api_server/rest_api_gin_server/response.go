package rest_api_gin_server

import "github.com/evgeniums/go-backend-helpers/pkg/api/api_server"

type Response struct {
	api_server.ResponseBase

	httpCode int
	request  *Request
}
