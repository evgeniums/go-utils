package app_with_multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pool_pubsub"
)

type AppWithMultitenancy interface {
	pool_pubsub.AppWithPubsub
	Multitenancy() multitenancy.Multitenancy
}

type AppWithMultitenancyBase struct {
	*pool_pubsub.AppWithPubsubBase
	tenancyManager *tenancy_manager.TenancyManager
}

func (a *AppWithMultitenancyBase) Multitenancy() multitenancy.Multitenancy {
	return a.tenancyManager
}

type MultitenancyConfigI interface {
	GetTenancyController() multitenancy.TenancyController
}

type MultitenancyConfig struct {
	TenancyController multitenancy.TenancyController
}

func (p *MultitenancyConfig) GetTenancyController() multitenancy.TenancyController {
	return p.TenancyController
}

type AppConfigI interface {
	pool_pubsub.AppConfigI
	MultitenancyConfigI
}

type AppConfig struct {
	pool_pubsub.AppConfig
	MultitenancyConfig
}

func NewApp(buildConfig *app_context.BuildConfig, tenancyDbModels []interface{}, appConfig ...AppConfigI) *AppWithMultitenancyBase {
	a := &AppWithMultitenancyBase{}
	if len(appConfig) == 0 {
		a.AppWithPubsubBase = pool_pubsub.NewApp(buildConfig)
		a.tenancyManager = tenancy_manager.NewTenancyManager(a.Pools(), a.Pubsub(), tenancyDbModels)
		a.tenancyManager.SetController(tenancy_manager.DefaultTenancyController(a.tenancyManager))
	} else {
		cfg := appConfig[0]
		a.AppWithPubsubBase = pool_pubsub.NewApp(buildConfig, cfg)
		a.tenancyManager = tenancy_manager.NewTenancyManager(a.Pools(), a.Pubsub(), tenancyDbModels)
		a.tenancyManager.SetController(cfg.GetTenancyController())
	}
	return a
}

func (a *AppWithMultitenancyBase) InitWithArgs(configFile string, args []string, configType ...string) (op_context.Context, error) {

	opCtx, err := a.AppWithPubsubBase.InitWithArgs(configFile, args, configType...)
	if err != nil {
		return nil, err
	}

	c := opCtx.TraceInMethod("AppWithMultitenancy.Init")
	defer opCtx.TraceOutMethod()

	err = a.tenancyManager.Init(opCtx, "multitenancy")
	if err != nil {
		msg := "failed to init multitenancy"
		c.SetMessage(msg)
		return opCtx, opCtx.Logger().PushFatalStack(msg, c.SetError(err))
	}

	return opCtx, nil
}

func (a *AppWithMultitenancyBase) Init(configFile string, configType ...string) (op_context.Context, error) {
	return a.InitWithArgs(configFile, nil, configType...)
}
