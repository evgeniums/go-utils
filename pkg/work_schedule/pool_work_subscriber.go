package work_schedule

import (
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pool_pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
)

type PoolWorkNotificationHandler[T Work] struct {
	pubsub_subscriber.SubscriberClientBase

	tenancies  multitenancy.Multitenancy
	controller *WorkSchedule[T]
}

func (p *PoolWorkNotificationHandler[T]) Handle(ctx op_context.Context, msg *PubsubWork[T]) error {

	c := ctx.TraceInMethod("PoolWorkNotificationHandler.Handle")
	defer ctx.TraceOutMethod()

	if msg.Immediate {

		var tenancy multitenancy.Tenancy
		var err error
		if msg.Tenancy != "" {
			tenancy, err = p.tenancies.Tenancy(msg.Tenancy)
			if err != nil {
				return c.SetError(err)
			}
		}

		p.controller.queue <- workItem[T]{work: msg.Work, tenancy: tenancy}
	}

	return nil
}

type PoolWorkSubscriber[T Work] struct {
	topic   *PubsubTopic[T]
	handler *PoolWorkNotificationHandler[T]
}

func NewPoolSubscriber[T Work](tenancies multitenancy.Multitenancy, controller *WorkSchedule[T]) *PoolWorkSubscriber[T] {
	p := &PoolWorkSubscriber[T]{}
	p.handler = &PoolWorkNotificationHandler[T]{
		tenancies:  tenancies,
		controller: controller,
	}
	return p
}

func (p *PoolWorkSubscriber[T]) Init(ctx op_context.Context, pubsub pool_pubsub.PoolPubsub) error {

	c := ctx.TraceInMethod("PoolWorkSubscriber.Init")
	defer ctx.TraceOutMethod()

	p.topic.TopicBase = pubsub_subscriber.New(PubsubTopicName, MakePubsubWork[T])
	_, err := pubsub.SubscribeSelfPool(ctx, p.topic)
	if err != nil {
		c.SetError(err)
		return ctx.Logger().PushFatalStack("failed to subscribe to pubsub notifications in self pool", err)
	}
	p.topic.Subscribe(p.handler)

	return nil
}
