package message_json

import (
	"encoding/json"
)

// JSON serializer.
type JsonSerializer struct {
}

func (j *JsonSerializer) ParseMessage(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

func (j *JsonSerializer) SerializeMessage(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}
