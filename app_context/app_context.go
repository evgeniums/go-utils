package app_context

import (
	"math/rand"
	"time"

	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/db"
	"github.com/evgeniums/go-backend-helpers/gorm_db"
	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/validator"
)

var Version = "development"
var Time = "unknown"
var Revision = "unknown"

type AppContext struct {
	logger.WithLoggerBase
	config.WithConfigBase

	GormDB            *gorm_db.GormDB
	ValidatorInstance *validator.PlaygroundValdator

	TestMode   bool
	TestValues map[string]interface{}
}

func (c *AppContext) DB() db.DB {
	return c.GormDB
}

func (c *AppContext) Validator() validator.Validator {
	return c.ValidatorInstance
}

func (c *AppContext) Testing() bool {
	return c.TestMode
}

func (c *AppContext) TestParameters() map[string]interface{} {
	return c.TestValues
}

func (c *AppContext) SetTestParameter(key string, value interface{}) {
	c.TestValues[key] = value
}

func (c *AppContext) GetTestParameter(key string) (interface{}, bool) {
	value, ok := c.TestValues[key]
	return value, ok
}

func NewAppContext() *AppContext {

	rand.Seed(time.Now().UTC().UnixNano())

	ctx := &AppContext{}
	ctx.ValidatorInstance = validator.NewPlaygroundValidator()
	return ctx
}

func (ctx *AppContext) Init(configFile string) error {

	// load configuration
	err := ctx.InitConfig(configFile)
	if err != nil {
		return err
	}

	// setup logger
	logConfigPath := "log"
	_ = ctx.InitLog(logConfigPath)
	l := ctx.Logger()
	l.Info("Starting...")

	// log build version
	l.Info("Build configuration", logger.Fields{"build_time": Time, "package_version": Version, "git_revision": Revision})

	// log app configuration
	ctx.LogConfigParameters(ctx.Logger(), "")
	ctx.LogConfigParameters(ctx.Logger(), logConfigPath)

	// connect to DB
	dbConfigPath := "psql"
	ctx.LogConfigParameters(ctx.Logger(), dbConfigPath)
	err = ctx.InitDb(dbConfigPath)
	if err != nil {
		return err
	}

	// setup testing
	ctx.TestMode = ctx.Config().GetBool("testing")
	if ctx.TestMode {
		ctx.Logger().Info("Running in test mode")
		ctx.TestValues = make(map[string]interface{})
	}

	return nil
}

func (ctx *AppContext) InitConfig(configFile string) error {
	ctx.ConfigInterface = &config.ConfigViper{}
	return ctx.Config().Init(configFile)
}

func (ctx *AppContext) InitLog(configPath string) error {

	log := &logger.LogrusLogger{}
	ctx.LoggerInterface = log

	var logConfig logger.LogrusConfig

	logConfig.Destination = ctx.Config().GetString(config.Key(configPath, "destination"))
	logConfig.File = ctx.Config().GetString(config.Key(configPath, "file"))
	logConfig.LogLevel = ctx.Config().GetString(config.Key(configPath, "level"))

	return log.Init(logConfig)
}

func (ctx *AppContext) InitDb(configPath string) error {

	ctx.GormDB = &gorm_db.GormDB{}
	ctx.GormDB.WithConfigBase = ctx.WithConfigBase
	ctx.GormDB.WithLoggerBase = ctx.WithLoggerBase

	return ctx.GormDB.Init(configPath)
}
