package pubsub_subscriber

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Subscriber interface {
	Subscribe(topic Topic) error
	Unsubscribe(topicName string)
	Shutdown(ctx context.Context) error

	Handle(ctx op_context.Context, topicName string, msg []byte) error
	Topic(topicName string) (Topic, error)
}

type WithSubscriber interface {
	Subscriber() Subscriber
}

type SubscriberBase struct {
	app_context.WithAppBase
	mutex      sync.RWMutex
	topics     map[string]Topic
	serializer message.Serializer
}

func (s *SubscriberBase) Construct(app app_context.Context, serializer ...message.Serializer) {
	s.WithAppBase.Init(app)
	s.topics = make(map[string]Topic)
	s.serializer = utils.OptionalArg(message.Serializer(message_json.Serializer), serializer...)
}

func (s *SubscriberBase) Topic(topicName string) (Topic, error) {
	s.mutex.RLock()
	topic, ok := s.topics[topicName]
	s.mutex.RUnlock()
	if !ok {
		return nil, errors.New("topic not found")
	}
	return topic, nil
}

func (s *SubscriberBase) Handle(ctx op_context.Context, topicName string, msg []byte) error {
	s.mutex.RLock()
	topic, ok := s.topics[topicName]
	s.mutex.RUnlock()
	if !ok {
		return nil
	}
	return topic.Handle(ctx, msg, s.serializer)
}

func (s *SubscriberBase) AddTopic(topic Topic) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.topics[topic.Name()]
	if exists {
		return errors.New("topic with such name already subscribed")
	}
	s.topics[topic.Name()] = topic
	return nil
}

func (s *SubscriberBase) DeleteTopic(topicName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.topics, topicName)
}

func (s *SubscriberBase) NewOpContext(topicName string) op_context.Context {

	opCtx := default_op_context.NewContext()
	opCtx.Init(s.App(), s.App().Logger(), s.App().Db())
	opCtx.SetName(topicName)
	errManager := &generic_error.ErrorManagerBase{}
	errManager.Init(http.StatusInternalServerError)
	opCtx.SetErrorManager(errManager)

	origin := default_op_context.NewOrigin(s.App())
	origin.SetUserType("pubsub")
	opCtx.SetOrigin(origin)

	return opCtx
}
