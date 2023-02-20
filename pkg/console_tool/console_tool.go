package console_tool

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/logger/logger_logrus"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
	"github.com/jessevdk/go-flags"
)

type MainOptions struct {
	ConfigFile   string `long:"config" description:"Configuration file"`
	ConfigFormat string `long:"config-format" description:"Format of configuration file" default:"json"`
	InitDb       string `long:"init-database" description:"Initialize database" default:"true"`
	DbSection    string `long:"database-section" description:"Database section in configuration file" default:"db"`
	InkokerName  string `long:"invoker-name" description:"Name of the user who invoked this utility"`
}

type Dummy struct{}
type ContextBulder = func(group string, command string) op_context.Context

type ConsoleUtility struct {
	Parser *flags.Parser
	App    app_context.Context
	Opts   MainOptions
	Args   []string

	InitApp func(config string) error
	InitDB  func() error
}

type InitApp = func(app app_context.Context, configFile string, args []string, configType ...string) error

type AppBuilder interface {
	NewApp(buildConfig *app_context.BuildConfig) app_context.Context
	InitApp(app app_context.Context, configFile string, args []string, configType ...string) error
}

func New(buildConfig *app_context.BuildConfig, appBuilder ...AppBuilder) *ConsoleUtility {
	c := &ConsoleUtility{}
	c.Parser = flags.NewParser(&c.Opts, flags.Default)
	var initApp InitApp
	if len(appBuilder) == 0 {
		c.App = app_default.New(buildConfig)
		initApp = func(app app_context.Context, configFile string, args []string, configType ...string) error {
			a := c.App.(*app_default.Context)
			return a.InitWithArgs(configFile, c.Args, c.Opts.ConfigFormat)
		}
	} else {
		c.App = appBuilder[0].NewApp(buildConfig)
		initApp = appBuilder[0].InitApp
	}
	c.InitApp = func(config string) error { return initApp(c.App, config, c.Args, c.Opts.ConfigFormat) }
	c.InitDB = func() error {
		initDbApp, ok := c.App.(app_default.WithInitGormDb)
		if !ok {
			return c.App.Logger().PushFatalStack("invalid type of application", errors.New("failed to cast application to app with initdb"))
		}
		return initDbApp.InitDB(c.Opts.DbSection)
	}
	return c
}

func (c *ConsoleUtility) Close() {
	c.App.Close()
}

func (c *ConsoleUtility) InitCommandContext(group string, command string) op_context.Context {
	err := c.InitApp(c.Opts.ConfigFile)
	if err != nil {
		app_context.AbortFatal(c.App, "failed to init application context", err)
	}

	if c.App.Cfg().GetString("logger.destination") != "stdout" {
		c.App.Cfg().Set("logger.destination", "stdout")
		consoleLogger := logger_logrus.New()
		err = consoleLogger.Init(c.App.Cfg(), c.App.Validator(), "logger")
		if err != nil {
			app_context.AbortFatal(c.App, "failed to init console logger", err)
		}
		teeLogger := logger.NewTee(c.App.Logger(), consoleLogger)
		c.App.SetLogger(teeLogger)
	}

	if c.Opts.InitDb == "true" {
		err = c.InitDB()
		if err != nil {
			app_context.AbortFatal(c.App, "failed to init database", err)
		}
	}

	opCtx := default_op_context.NewContext()
	opCtx.Init(c.App, c.App.Logger(), c.App.Db())
	opCtx.SetName(fmt.Sprintf("%s.%s", group, command))
	errManager := &generic_error.ErrorManagerBase{}
	errManager.Init(http.StatusBadRequest)
	opCtx.SetErrorManager(errManager)

	origin := default_op_context.NewOrigin(c.App)
	origin.SetUser(c.Opts.InkokerName)
	origin.SetUserType("console")
	opCtx.SetOrigin(origin)

	return opCtx
}

func (c *ConsoleUtility) Parse() {
	var err error
	c.Args, err = c.Parser.Parse()
	if err != nil {
		if err, ok := err.(*flags.Error); ok {
			if err.Type == flags.ErrHelp {
				os.Exit(0)
			}
			c.Parser.WriteHelp(os.Stdout)
		}
		os.Exit(1)
	}
}

func (c *ConsoleUtility) AddCommand(handler func(ctxBuilder ContextBulder, parser *flags.Parser)) {
	handler(c.InitCommandContext, c.Parser)
}
