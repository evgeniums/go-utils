package rest_api_gin_server

import "github.com/evgeniums/go-backend-helpers/pkg/api_server"

type Response struct {
	api_server.ResponseBase

	httpCode int
	request  *Request
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}
