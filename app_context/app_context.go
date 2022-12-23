package app_context

import (
	"math/rand"
	"time"

	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/config/config_viper"
	"github.com/evgeniums/go-backend-helpers/db"
	"github.com/evgeniums/go-backend-helpers/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/logger/logger_logrus"
	"github.com/evgeniums/go-backend-helpers/validator"
	"github.com/evgeniums/go-backend-helpers/validator/validator_playground"
)

var Version = "development"
var Time = "unknown"
var Revision = "unknown"

type AppContext struct {
	logger.WithLoggerBase
	config.WithConfigBase

	GormDB            *db_gorm.GormDB
	ValidatorInstance *validator_playground.PlaygroundValdator

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
	ctx.ValidatorInstance = validator_playground.New()
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
	ctx.ConfigInterface = config_viper.New()
	return ctx.Config().Init(configFile)
}

func (ctx *AppContext) InitLog(configPath string) error {

	log := logger_logrus.New()
	ctx.LoggerInterface = log

	return log.Init(ctx.Config(), "logger")
}

func (ctx *AppContext) InitDb(configPath string) error {

	ctx.GormDB = &db_gorm.GormDB{}
	ctx.GormDB.WithConfigBase = ctx.WithConfigBase
	ctx.GormDB.WithLoggerBase = ctx.WithLoggerBase

	return ctx.GormDB.Init(configPath)
}
