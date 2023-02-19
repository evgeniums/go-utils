package pool_pubsub

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
)

type PoolPubsub interface {
	Shutdown()

	PublishSelfPool(topicName string, msg interface{}) error
	PublishPools(topicName string, msg interface{}, poolIds ...string) error

	SubscribeSelfPool(topic pubsub_subscriber.Topic) error
	UnsubscribeSelfPool(topicName string)
	SubscriberTopicInSelfPool(topicName string) (pubsub_subscriber.Topic, error)
}

type PoolPubsubBase struct {
	factory            pubsub_factory.PubsubFactory
	selfPoolSubscriber pubsub_subscriber.Subscriber
	publishers         map[string]pubsub.Publisher
	selfPoolPublisher  pubsub.Publisher
}

func New(factory ...pubsub_factory.PubsubFactory) *PoolPubsubBase {
	p := &PoolPubsubBase{}
	if len(factory) == 0 {
		p.factory = pubsub_factory.DefaultPubsubFactory()
	} else {
		p.factory = factory[0]
	}
	return p
}

func (p *PoolPubsubBase) Init(app app_context.Context, pools pool.PoolStore) error {

	makePublisher := func(poo pool.Pool) (pubsub.Publisher, *pubsub_factory.PubsubConfig, error) {
		service, err := poo.Service(pool.TypePubsub)
		if err != nil {
			return nil, nil, app.Logger().PushFatalStack("failed to find pubsub service in self pool", err)
		}
		cfg := &pubsub_factory.PubsubConfig{PoolService: service}
		publisher, err := p.factory.MakePublisher(app, cfg)
		if err != nil {
			return nil, nil, app.Logger().PushFatalStack("failed to make pubsub publisher in self pool", err)
		}
		p.publishers[poo.GetID()] = publisher
		return publisher, cfg, nil
	}

	selfPool, err := pools.SelfPool()
	if err != nil {
		var cfg *pubsub_factory.PubsubConfig
		p.selfPoolPublisher, cfg, err = makePublisher(selfPool)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub publisher in self pool", err)
		}
		p.selfPoolSubscriber, err = p.factory.MakeSubscriber(app, cfg)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub subscriber in self pool", err)
		}
		return nil
	}

	for _, pool := range pools.Pools() {
		_, _, err := makePublisher(pool)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub publisher in self pool", err, logger.Fields{"pool_name": pool.Name(), "pool_id": pool.GetID()})
		}
	}

	return nil
}

func (p *PoolPubsubBase) Shutdown() {
	if p.selfPoolPublisher != nil {
		p.selfPoolPublisher.Shutdown()
	}
	if p.selfPoolSubscriber != nil {
		p.selfPoolSubscriber.Shutdown()
	}
	for _, publisher := range p.publishers {
		publisher.Shutdown()
	}
}

func (p *PoolPubsubBase) PublishSelfPool(topicName string, msg interface{}) error {
	if p.selfPoolPublisher == nil {
		return errors.New("self publisher not set")
	}
	return p.selfPoolPublisher.Publish(topicName, msg)
}

func (p *PoolPubsubBase) PublishPools(topicName string, msg interface{}, poolIds ...string) error {
	if len(poolIds) == 0 {
		// publish to all pools
		for poolId, publisher := range p.publishers {
			err := publisher.Publish(topicName, msg)
			if err != nil {
				return fmt.Errorf("failed to publish to %s pool", poolId)
			}
		}
	} else {
		// publish to specific pools
		for _, poolId := range poolIds {
			publisher, ok := p.publishers[poolId]
			if ok {
				err := publisher.Publish(topicName, msg)
				if err != nil {
					return fmt.Errorf("failed to publish to %s pool", poolId)
				}
			}
		}
	}
	return nil
}

func (p *PoolPubsubBase) SubscribeSelfPool(topic pubsub_subscriber.Topic) error {
	if p.selfPoolSubscriber == nil {
		return errors.New("self pool subscriber not set")
	}
	return p.selfPoolSubscriber.Subscribe(topic)
}

func (p *PoolPubsubBase) UnsubscribeSelfPool(topicName string) {
	if p.selfPoolSubscriber != nil {
		p.selfPoolSubscriber.Unsubscribe(topicName)
	}
}

func (p *PoolPubsubBase) SubscriberTopicInSelfPool(topicName string) (pubsub_subscriber.Topic, error) {
	if p.selfPoolSubscriber == nil {
		return nil, errors.New("self pool subscriber not set")
	}
	return p.selfPoolSubscriber.Topic(topicName)
}
