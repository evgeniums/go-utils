package pool_pubsub

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
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

	SubscribePools(topic pubsub_subscriber.Topic, poolIds ...string) error
	UnsubscribePools(topicName string, poolIds ...string)
	SubscriberTopicInPool(topicName string, poolId string) (pubsub_subscriber.Topic, error)
}

type PoolPubsubBase struct {
	factory            pubsub_factory.PubsubFactory
	selfPoolSubscriber pubsub_subscriber.Subscriber
	publishers         map[string]pubsub.Publisher
	selfPoolPublisher  pubsub.Publisher
	subscribers        map[string]pubsub_subscriber.Subscriber
}

func NewPubsub(factory ...pubsub_factory.PubsubFactory) *PoolPubsubBase {
	p := &PoolPubsubBase{}
	if len(factory) != 0 {
		p.factory = factory[0]
	}
	if p.factory == nil {
		p.factory = pubsub_factory.DefaultPubsubFactory()
	}
	p.subscribers = make(map[string]pubsub_subscriber.Subscriber)
	p.publishers = make(map[string]pubsub.Publisher)
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
			return nil, nil, app.Logger().PushFatalStack("failed to make pubsub publisher for pool", err, db.Fields{"pool_id": poo.GetID(), "pool_name": poo.Name()})
		}
		p.publishers[poo.GetID()] = publisher
		return publisher, cfg, nil
	}

	makeSubscriber := func(poo pool.Pool, cfg *pubsub_factory.PubsubConfig) (pubsub_subscriber.Subscriber, error) {
		subscriber, err := p.factory.MakeSubscriber(app, cfg)
		if err != nil {
			return nil, app.Logger().PushFatalStack("failed to make pubsub subscriber for pool", err, db.Fields{"pool_id": poo.GetID(), "pool_name": poo.Name()})
		}
		p.subscribers[poo.GetID()] = subscriber
		return subscriber, nil
	}

	selfPool, err := pools.SelfPool()
	if err != nil {
		var cfg *pubsub_factory.PubsubConfig
		p.selfPoolPublisher, cfg, err = makePublisher(selfPool)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub publisher in self pool", err)
		}
		p.selfPoolSubscriber, err = makeSubscriber(selfPool, cfg)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub subscriber in self pool", err)
		}
		return nil
	}

	for _, pool := range pools.Pools() {
		_, cfg, err := makePublisher(pool)
		if err != nil {
			return err
		}
		_, err = makeSubscriber(pool, cfg)
		if err != nil {
			return err
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

func (p *PoolPubsubBase) SubscribePools(topic pubsub_subscriber.Topic, poolIds ...string) error {
	if len(poolIds) == 0 {
		// subscribe to all pools
		for poolId, subscriber := range p.subscribers {
			err := subscriber.Subscribe(topic)
			if err != nil {
				return fmt.Errorf("failed to subscribe to %s pool", poolId)
			}
		}
	} else {
		// subscribe to specific pools
		for _, poolId := range poolIds {
			subscriber, ok := p.subscribers[poolId]
			if ok {
				err := subscriber.Subscribe(topic)
				if err != nil {
					return fmt.Errorf("failed to subscribe to %s pool", poolId)
				}
			}
		}
	}
	return nil
}

func (p *PoolPubsubBase) UnsubscribePools(topicName string, poolIds ...string) {
	if len(poolIds) == 0 {
		// unsubscribe from all pools
		for _, subscriber := range p.subscribers {
			subscriber.Unsubscribe(topicName)
		}
	} else {
		// unsubscribe from specific pools
		for _, poolId := range poolIds {
			subscriber, ok := p.subscribers[poolId]
			if ok {
				subscriber.Unsubscribe(topicName)
			}
		}
	}
}

func (p *PoolPubsubBase) SubscriberTopicInPool(topicName string, poolId string) (pubsub_subscriber.Topic, error) {
	subscriber, ok := p.subscribers[poolId]
	if !ok {
		return nil, errors.New("no subscriber for that pool")
	}
	return subscriber.Topic(topicName)
}
