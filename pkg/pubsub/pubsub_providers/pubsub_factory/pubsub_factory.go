package pubsub_factory

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_inmem"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_redis"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type PubsubConfigI interface {
	GetPoolService() *pool.PoolServiceBinding
	GetConfigKeyPath() string
}

type PubsubConfig struct {
	PoolService   *pool.PoolServiceBinding
	ConfigKeyPath string
}

func (p *PubsubConfig) GetPoolService() *pool.PoolServiceBinding {
	return p.PoolService
}

func (p *PubsubConfig) GetConfigKeyPath() string {
	return p.ConfigKeyPath
}

type PubsubFactory interface {
	MakePublisher(app app_context.Context, config ...PubsubConfigI) (pubsub.Publisher, error)
	MakeSubscriber(app app_context.Context, config ...PubsubConfigI) (pubsub_subscriber.Subscriber, error)
}

type PubsubFactoryBase struct {
	serializer message.Serializer
	inmem      *pubsub_inmem.PubsubInmem
}

func splitConfig(app app_context.Context, config ...PubsubConfigI) (*pool.PoolServiceBinding, string) {
	configPath := "pubsub"
	var poolService *pool.PoolServiceBinding
	if len(config) != 0 {
		configPath = config[0].GetConfigKeyPath()
		poolService = config[0].GetPoolService()
	}
	return poolService, configPath
}

func provider(app app_context.Context, poolService *pool.PoolServiceBinding, configPath string) string {

	if poolService != nil {
		return poolService.Provider()
	}

	providerKey := object_config.Key(configPath, "provider")
	provider := app.Cfg().GetString(providerKey)
	return provider
}

func (p *PubsubFactoryBase) MakeInmemPubsub(app app_context.Context) (*pubsub_inmem.PubsubInmem, error) {

	if p.inmem != nil {
		return p.inmem, nil
	}

	p.inmem = pubsub_inmem.New(app, p.serializer)
	return p.inmem, nil
}

func (p *PubsubFactoryBase) MakePublisher(app app_context.Context, config ...PubsubConfigI) (pubsub.Publisher, error) {

	poolService, configPath := splitConfig(app, config...)
	provider := provider(app, poolService, configPath)
	if provider == pubsub_redis.Provider {
		publisher := pubsub_redis.NewPublisher(p.serializer)
		return publisher, nil
	} else if provider == pubsub_inmem.Provider {
		return p.MakeInmemPubsub(app)
	}

	return nil, errors.New("unknown provider")
}

func (p *PubsubFactoryBase) MakeSubscriber(app app_context.Context, config ...PubsubConfigI) (pubsub_subscriber.Subscriber, error) {

	poolService, configPath := splitConfig(app, config...)
	provider := provider(app, poolService, configPath)

	if provider == pubsub_redis.Provider {
		subsciber := pubsub_redis.NewSubscriber(app, p.serializer)
		var err error

		if poolService == nil {
			err = subsciber.Init(app.Cfg(), app.Logger(), app.Validator(), configPath)
		} else {
			cfg := &pubsub_redis.RedisConfig{}
			cfg.Host = poolService.PrivateHost()
			cfg.Port = poolService.PrivatePort()
			cfg.Password = poolService.Secret1()
			db := 0
			if poolService.Parameter1() != "" {
				dbU, err := utils.StrToUint32(poolService.Parameter1())
				if err != nil {
					return nil, errors.New("invalid number of redis database")
				}
				db = int(dbU)
			}
			cfg.Db = db
			err = subsciber.InitWithConfig(app.Logger(), cfg)
		}

		if err != nil {
			return nil, err
		}
	} else if provider == pubsub_inmem.Provider {
		return p.MakeInmemPubsub(app)
	}

	return nil, errors.New("unknown provider")
}

func DefaultPubsubFactory(serializer ...message.Serializer) PubsubFactory {
	f := &PubsubFactoryBase{}
	f.serializer = utils.OptionalArg(message.Serializer(message_json.Serializer), serializer...)
	return f
}
