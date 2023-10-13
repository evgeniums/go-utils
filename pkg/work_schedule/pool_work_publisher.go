package work_schedule

import (
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pool_pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type PubsubWork[T Work] struct {
	Work      T      `json:"work"`
	Immediate bool   `json:"immediate"`
	Tenancy   string `json:"tenancy"`
}

const PubsubTopicName = "work_schedule"

type PubsubTopic[T Work] struct {
	*pubsub_subscriber.TopicBase[*PubsubWork[T]]
}

func MakePubsubWork[T Work]() *PubsubWork[T] {
	return &PubsubWork[T]{}
}

func NewPubsubWork[T Work](work T, immediate bool, tenancy ...multitenancy.Tenancy) *PubsubWork[T] {
	w := &PubsubWork[T]{Work: work, Immediate: immediate}
	t := utils.OptionalArg(nil, tenancy...)
	if t != nil {
		w.Tenancy = t.GetID()
	}
	return w
}

type PoolWorkPublisher[T Work] struct {
	pubsub pool_pubsub.PoolPubsub
}

func NewPoolWorkPublisher[T Work](pubsub pool_pubsub.PoolPubsub) *PoolWorkPublisher[T] {
	p := &PoolWorkPublisher[T]{pubsub: pubsub}
	return p
}

func (p *PoolWorkPublisher[T]) PostWork(ctx op_context.Context, work T, immediate bool, tenancy ...multitenancy.Tenancy) error {

	c := ctx.TraceInMethod("PoolWorkPublisher.PostWork")
	defer ctx.TraceOutMethod()

	msg := NewPubsubWork(work, immediate, tenancy...)
	err := p.pubsub.PublishSelfPool(PubsubTopicName, msg)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}
