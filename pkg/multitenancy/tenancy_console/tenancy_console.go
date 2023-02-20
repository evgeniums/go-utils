package tenancy_console

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type MultitenancyAppBuilder struct {
	Models []interface{}
	Config app_with_multitenancy.AppConfigI
	App    *app_with_multitenancy.AppWithMultitenancyBase
}

func (m *MultitenancyAppBuilder) NewApp(buildConfig *app_context.BuildConfig) app_context.Context {
	if m.Config == nil {
		m.App = app_with_multitenancy.NewApp(buildConfig, m.Models)
	} else {
		m.App = app_with_multitenancy.NewApp(buildConfig, m.Models, m.Config)
	}
	return m.App
}

func (m *MultitenancyAppBuilder) InitApp(app app_context.Context, configFile string, args []string, configType ...string) error {
	a, ok := app.(*app_with_multitenancy.AppWithMultitenancyBase)
	if !ok {
		return app.Logger().PushFatalStack("invalid application type", errors.New("failed to cast app to multitenancy app"))
	}
	ctx, err := a.InitWithArgs(configFile, args, configType...)
	ctx.Close()
	return err
}

type AppBuilder interface {
	NewApp() app_context.Context
	InitApp(app app_context.Context, configFile string, args []string, configType ...string) error
}

type TenancyCommands struct {
	console_tool.Commands[*TenancyCommands]
	MakeController func(app app_context.Context) multitenancy.TenancyController
}

func NewTenancyCommands() *TenancyCommands {
	p := &TenancyCommands{}
	p.Construct(p, "tenancy", "Manage tenancies")
	p.LoadHandlers()
	return p
}

func (p *TenancyCommands) LoadHandlers() {
	p.AddHandlers(Add,
		Find,
		List,
		Activate,
		Deactivate,
		Customer,
		Role,
		Path,
		ChangePoolOrDb,
		Delete,
	)
}

type Handler = console_tool.Handler[*TenancyCommands]

type HandlerBase struct {
	console_tool.HandlerBase[*TenancyCommands]
}

func (b *HandlerBase) Context() (op_context.Context, multitenancy.TenancyController) {
	ctx := b.HandlerBase.Context()

	a, ok := ctx.App().(*app_with_multitenancy.AppWithMultitenancyBase)
	if !ok {
		panic(fmt.Errorf("invalid application type: %s", errors.New("failed to cast app to multitenancy app")))
	}

	return ctx, a.Multitenancy().TenancyController()
}
