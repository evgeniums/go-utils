package message_json

import (
	"encoding/json"

	"github.com/evgeniums/go-backend-helpers/pkg/message"
)

// Base type for containers of JSON messages.
type WithMessageJson struct {
	message.WithMessageBase
}

func (w *WithMessageJson) ParseMessage(data []byte) error {
	return json.Unmarshal(data, w.Message())
}

func (w *WithMessageJson) SerializeMessage() ([]byte, error) {
	return json.Marshal(w.Message())
}
