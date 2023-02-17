package pubsub

import (
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
)

type Subscriber interface {
	Subscribe(topic Topic) error
	Unsubscribe(topicName string)
	Shutdown()

	Handle(ctx op_context.Context, topicName string, msg []byte) error
	Topic(topicName string) (Topic, error)
}

type SubscriberBase struct {
	app_context.WithAppBase
	mutex  sync.RWMutex
	topics map[string]Topic
}

func (s *SubscriberBase) Init(app app_context.Context) {
	s.WithAppBase.Init(app)
	s.topics = make(map[string]Topic)
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
	return topic.Handle(ctx, msg)
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

	opCtx := &default_op_context.ContextBase{}
	opCtx.Init(s.App(), s.App().Logger(), s.App().Db())
	opCtx.SetName(topicName)
	errManager := &generic_error.ErrorManagerBase{}
	errManager.Init(http.StatusInternalServerError)
	opCtx.SetErrorManager(errManager)

	origin := &default_op_context.Origin{}
	origin.SetType(s.App().Application())
	origin.SetName(s.App().AppInstance())
	hostname, _ := os.Hostname()
	origin.SetSource(hostname)
	origin.SetUserType("pubsub")
	opCtx.SetOrigin(origin)

	return opCtx
}
