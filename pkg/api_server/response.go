package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/message"
)

// Interface of response of server API.
type Response interface {
	message.WithMessage
}
