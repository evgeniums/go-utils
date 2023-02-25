package pool_pubsub

import (
	"context"
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
	Shutdown(ctx context.Context) error

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

		fields := db.Fields{"pool_id": poo.GetID(), "pool_name": poo.Name()}

		service, err := poo.Service(pool.TypePubsub)
		if err != nil {
			return nil, nil, app.Logger().PushFatalStack("failed to find pubsub service in pool", err, fields)
		}
		if !service.IsActive() {
			fields["service_name"] = service.ServiceName
			app.Logger().Warn("Pubsub skipped for inactive service", fields)
			return nil, nil, nil
		}
		cfg := &pubsub_factory.PubsubConfig{PoolService: service}
		publisher, err := p.factory.MakePublisher(app, cfg)
		if err != nil {
			return nil, nil, app.Logger().PushFatalStack("failed to make pubsub publisher for pool", err, fields)
		}
		p.publishers[poo.GetID()] = publisher
		app.Logger().Info("Pubsub publisher connected", fields)
		return publisher, cfg, nil
	}

	makeSubscriber := func(poo pool.Pool, cfg *pubsub_factory.PubsubConfig) (pubsub_subscriber.Subscriber, error) {

		fields := db.Fields{"pool_id": poo.GetID(), "pool_name": poo.Name()}

		subscriber, err := p.factory.MakeSubscriber(app, cfg)
		if err != nil {
			return nil, app.Logger().PushFatalStack("failed to make pubsub subscriber for pool", err, fields)
		}
		p.subscribers[poo.GetID()] = subscriber
		app.Logger().Info("Pubsub subscriber connected", fields)
		return subscriber, nil
	}

	selfPool, err := pools.SelfPool()
	if err == nil {

		if !selfPool.IsActive() {
			fields := db.Fields{"pool_id": selfPool.GetID(), "pool_name": selfPool.Name()}
			app.Logger().Warn("Pubsub skipped for inactive pool", fields)
			return nil
		}

		var cfg *pubsub_factory.PubsubConfig
		p.selfPoolPublisher, cfg, err = makePublisher(selfPool)
		if err != nil {
			return app.Logger().PushFatalStack("failed to make pubsub publisher in self pool", err)
		}
		if cfg != nil {
			p.selfPoolSubscriber, err = makeSubscriber(selfPool, cfg)
			if err != nil {
				return app.Logger().PushFatalStack("failed to make pubsub subscriber in self pool", err)
			}
		}
		return nil
	}

	for _, pool := range pools.Pools() {
		if pool.IsActive() {
			_, cfg, err := makePublisher(pool)
			if err != nil {
				return err
			}
			if cfg != nil {
				_, err = makeSubscriber(pool, cfg)
				if err != nil {
					return err
				}
			}
		} else {
			fields := db.Fields{"pool_id": pool.GetID(), "pool_name": pool.Name()}
			app.Logger().Warn("Pubsub skipped for inactive pool", fields)
		}
	}

	return nil
}

func (p *PoolPubsubBase) Shutdown(ctx context.Context) error {
	var err error
	if p.selfPoolPublisher != nil {
		err1 := p.selfPoolPublisher.Shutdown(ctx)
		if err1 != nil {
			err = err1
		}
	}
	if p.selfPoolSubscriber != nil {
		err1 := p.selfPoolSubscriber.Shutdown(ctx)
		if err1 != nil {
			err = err1
		}
	}
	for _, publisher := range p.publishers {
		err1 := publisher.Shutdown(ctx)
		if err1 != nil {
			err = err1
		}
	}
	return err
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
