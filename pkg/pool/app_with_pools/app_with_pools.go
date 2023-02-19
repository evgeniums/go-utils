package app_with_pools

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
)

type AppWithPools interface {
	app_context.Context
	Pools() pool.PoolStore
}

type AppWithPoolsBase struct {
	*app_default.Context
	pools *pool.PoolStoreBase
}

func (a *AppWithPoolsBase) Pools() pool.PoolStore {
	return a.pools
}

type AppConfigI interface {
	app_default.AppConfigI
	pool.PoolStoreConfigI
}

type AppConfig struct {
	app_default.AppConfig
	pool.PoolStoreConfig
}

func New(buildConfig *app_context.BuildConfig, appConfig ...AppConfigI) *AppWithPoolsBase {
	a := &AppWithPoolsBase{}
	if len(appConfig) == 0 {
		a.Context = app_default.New(buildConfig)
		a.pools = pool.NewPoolStore()
	} else {
		cfg := appConfig[0]
		a.Context = app_default.New(buildConfig, cfg)
		a.pools = pool.NewPoolStore(cfg)
	}
	return a
}

func (a *AppWithPoolsBase) InitWithArgs(configFile string, args []string, configType ...string) (op_context.Context, error) {

	err := a.Context.InitWithArgs(configFile, args, configType...)
	if err != nil {
		return nil, err
	}

	opCtx := default_op_context.NewAppInitContext(a)
	c := opCtx.TraceInMethod("AppWithPools.Init")
	defer opCtx.TraceOutMethod()

	err = a.pools.Init(opCtx, "pools")
	if err != nil {
		msg := "failed to init pools"
		c.SetMessage(msg)
		return opCtx, opCtx.Logger().PushFatalStack(msg, c.SetError(err))
	}

	return opCtx, nil
}

func (a *AppWithPoolsBase) Init(configFile string, configType ...string) (op_context.Context, error) {
	return a.InitWithArgs(configFile, nil, configType...)
}
