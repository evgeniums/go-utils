package pubsub_inmem

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
)

type PubsubInmem struct {
	pubsub.SubscriberBase
	pubsub.PublisherBase
}

func New(app app_context.Context, serializer ...message.Serializer) *PubsubInmem {
	p := &PubsubInmem{}
	p.PublisherBase.Construct(serializer...)
	p.SubscriberBase.Construct(app, serializer...)
	return p
}

func (p *PubsubInmem) Shutdown() {
}

func (p *PubsubInmem) Subscribe(topic pubsub.Topic) error {
	return p.AddTopic(topic)
}

func (p *PubsubInmem) Unsubscribe(topicName string) {
	p.DeleteTopic(topicName)
}

func (p *PubsubInmem) Publish(topicName string, obj interface{}) error {

	msg, err := p.Serialize(obj)
	if err != nil {
		return err
	}

	opCtx := p.NewOpContext(topicName)
	defer opCtx.Close()

	p.Handle(opCtx, topicName, msg)

	return nil
}
