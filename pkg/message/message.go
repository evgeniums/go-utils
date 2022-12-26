package message

// Base interface for containers with messages.
type WithMessage interface {
	SetMessage(message interface{})
	Message() interface{}
	ParseMessage(data []byte) error
	SerializeMessage() ([]byte, error)
}

// Base type for containers with messages.
type WithMessageBase struct {
	message interface{}
}

func (w *WithMessageBase) SetMessage(message interface{}) {
	w.message = message
}

func (w *WithMessageBase) Message() interface{} {
	return w.message
}
