package pubsub

import (
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type SubscriberClient[T any] interface {
	Id() string
	Handle(ctx op_context.Context, msg T) error
}

type SubscriberClientBase struct {
	id string
}

func (s *SubscriberClientBase) Init(id string) {
	s.id = id
}

func (s *SubscriberClientBase) Id() string {
	return s.id
}

type Topic interface {
	Name() string
	Handle(ctx op_context.Context, msg []byte) error
	Unsubscribe(id string)
}

type TopicT[T any] interface {
	Topic
	Subscribe(subscriber SubscriberClient[T])
}

type TopicBase[T any] struct {
	mutex        sync.RWMutex
	name         string
	subscribers  map[string]SubscriberClient[T]
	builder      func() T
	deserializer message.Serializer
}

func New[T any](name string, builder func() T, deserializer ...message.Serializer) *TopicBase[T] {
	t := &TopicBase[T]{}
	t.name = name
	t.builder = builder
	t.subscribers = make(map[string]SubscriberClient[T])
	t.deserializer = utils.OptionalArg(message.Serializer(message_json.Serializer), deserializer...)
	return t
}

func (t *TopicBase[T]) Name() string {
	return t.name
}

func (t *TopicBase[T]) Unsubscribe(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.subscribers, id)
}

func (t *TopicBase[T]) Subscribe(subscriber SubscriberClient[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.subscribers[subscriber.Id()] = subscriber
}

func (t *TopicBase[T]) Handle(ctx op_context.Context, msg []byte) error {

	c := ctx.TraceInMethod("pubsub.Topic.Handle")
	defer ctx.TraceOutMethod()

	obj := t.builder()
	err := t.deserializer.ParseMessage(msg, obj)
	if err != nil {
		c.SetMessage("failed to unmarshal message")
		return c.SetError(err)
	}

	t.mutex.RLock()
	subscribers := utils.AllMapValues(t.subscribers)
	t.mutex.RUnlock()

	for _, subscriber := range subscribers {
		err = subscriber.Handle(ctx, obj)
		if err != nil {
			c.Logger().Warn("failed to handle message", logger.Fields{"subscriber": subscriber.Id()})
		}
	}

	return nil
}
