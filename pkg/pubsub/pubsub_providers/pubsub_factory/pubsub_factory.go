package pubsub_factory

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_inmem"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_redis"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type PubsubFactory interface {
	MakePublisher(app app_context.Context, configPath ...string) (pubsub.Publisher, error)
	MakeSubscriber(app app_context.Context, configPath ...string) (pubsub_subscriber.Subscriber, error)
}

type PubsubFactoryBase struct {
	serializer message.Serializer
	inmem      *pubsub_inmem.PubsubInmem
}

func provider(app app_context.Context, configPath ...string) string {
	config := utils.OptionalArg("pubsub", configPath...)
	providerKey := object_config.Key(config, "provider")
	provider := app.Cfg().GetString(providerKey)
	return provider
}

func (p *PubsubFactoryBase) MakePublisher(app app_context.Context, configPath ...string) (pubsub.Publisher, error) {
	provider := provider(app, configPath...)
	if provider == "redis" {
		publisher := pubsub_redis.NewPublisher(p.serializer)
		return publisher, nil
	}

	if p.inmem != nil {
		return p.inmem, nil
	}

	p.inmem = pubsub_inmem.New(app, p.serializer)
	return p.inmem, nil
}

func (p *PubsubFactoryBase) MakeSubscriber(app app_context.Context, configPath ...string) (pubsub_subscriber.Subscriber, error) {
	provider := provider(app, configPath...)
	if provider == "redis" {
		subsciber := pubsub_redis.NewSubscriber(app, p.serializer)
		err := subsciber.Init(app.Cfg(), app.Logger(), app.Validator(), configPath...)
		if err != nil {
			return nil, err
		}
	}

	if p.inmem != nil {
		return p.inmem, nil
	}

	p.inmem = pubsub_inmem.New(app, p.serializer)
	return p.inmem, nil
}

func DefaultPubsubFactory(serializer ...message.Serializer) PubsubFactory {
	f := &PubsubFactoryBase{}
	f.serializer = utils.OptionalArg(message.Serializer(message_json.Serializer), serializer...)
	return f
}
