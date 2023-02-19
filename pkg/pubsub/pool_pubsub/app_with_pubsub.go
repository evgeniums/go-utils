package pool_pubsub

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool/app_with_pools"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_factory"
)

type AppWithPubsub interface {
	app_with_pools.AppWithPools
	Pubsub() PoolPubsub
}

type AppWithPubsubBase struct {
	*app_with_pools.AppWithPoolsBase
	pubsub *PoolPubsubBase
}

func (a *AppWithPubsubBase) Pubsub() PoolPubsub {
	return a.pubsub
}

type PoolPubsubConfigI interface {
	GetPubsubFactory() pubsub_factory.PubsubFactory
}

type PoolPubsubConfig struct {
	PubsubFactory pubsub_factory.PubsubFactory
}

func (p *PoolPubsubConfig) GetPubsubFactory() pubsub_factory.PubsubFactory {
	return p.PubsubFactory
}

type AppConfigI interface {
	app_with_pools.AppConfigI
	PoolPubsubConfigI
}

type AppConfig struct {
	app_with_pools.AppConfig
	PoolPubsubConfig
}

func NewApp(buildConfig *app_context.BuildConfig, appConfig ...AppConfigI) *AppWithPubsubBase {
	a := &AppWithPubsubBase{}
	if len(appConfig) == 0 {
		a.AppWithPoolsBase = app_with_pools.New(buildConfig)
		a.pubsub = NewPubsub()
	} else {
		cfg := appConfig[0]
		a.AppWithPoolsBase = app_with_pools.New(buildConfig, cfg)
		a.pubsub = NewPubsub(cfg.GetPubsubFactory())
	}
	return a
}

func (a *AppWithPubsubBase) InitWithArgs(configFile string, args []string, configType ...string) (op_context.Context, error) {

	opCtx, err := a.AppWithPoolsBase.InitWithArgs(configFile, args, configType...)
	if err != nil {
		return nil, err
	}

	c := opCtx.TraceInMethod("AppWithPubsub.Init")
	defer opCtx.TraceOutMethod()

	err = a.pubsub.Init(a, a.Pools())
	if err != nil {
		msg := "failed to init pubsub"
		c.SetMessage(msg)
		return opCtx, opCtx.Logger().PushFatalStack(msg, c.SetError(err))
	}

	return opCtx, nil
}

func (a *AppWithPubsubBase) Init(configFile string, configType ...string) (op_context.Context, error) {
	return a.InitWithArgs(configFile, nil, configType...)
}
