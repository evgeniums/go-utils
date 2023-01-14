package rest_api_gin_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
)

type Response struct {
	message_json.WithMessageJson

	httpCode int
	request  *Request
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}
