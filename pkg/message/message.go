package message

// Base interface for containers with messages.
type Serializer interface {
	ParseMessage(data []byte, message interface{}) error
	SerializeMessage(message interface{}) ([]byte, error)
	Format() string
	ContentMime() string
}
