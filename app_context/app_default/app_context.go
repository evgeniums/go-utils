package app_default

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

type ContextConfig struct {
	testing    bool
	testValues map[string]interface{}
}

type Context struct {
	logger.WithLoggerBase
	config.WithConfigBase

	db        *db_gorm.GormDB
	validator *validator_playground.PlaygroundValdator

	ContextConfig
}

func (c *Context) DB() db.DB {
	return c.db
}

func (c *Context) Validator() validator.Validator {
	return c.validator
}

func (c *Context) Testing() bool {
	return c.testing
}

func (c *Context) TestParameters() map[string]interface{} {
	return c.testValues
}

func (c *Context) SetTestParameter(key string, value interface{}) {
	c.testValues[key] = value
}

func (c *Context) GetTestParameter(key string) (interface{}, bool) {
	value, ok := c.testValues[key]
	return value, ok
}

func New() *Context {

	rand.Seed(time.Now().UTC().UnixNano())

	c := &Context{}
	c.validator = validator_playground.New()

	return c
}

func (c *Context) Init(configFile string) error {

	// load configuration
	err := c.InitConfig(configFile)
	if err != nil {
		return err
	}

	// setup logger
	logConfigPath := "log"
	_ = c.InitLog(logConfigPath)
	l := c.Logger()
	l.Info("Starting...")

	// log build version
	l.Info("Build configuration", logger.Fields{"build_time": Time, "package_version": Version, "git_revision": Revision})

	// log app configuration
	c.LogConfigParameters(c.Logger(), "")
	c.LogConfigParameters(c.Logger(), logConfigPath)

	// connect to DB
	dbConfigPath := "psql"
	c.LogConfigParameters(c.Logger(), dbConfigPath)
	err = c.InitDb(dbConfigPath)
	if err != nil {
		return err
	}

	// setup testing
	c.testing = c.Config().GetBool("testing")
	if c.Testing() {
		c.Logger().Info("Running in test mode")
		c.testValues = make(map[string]interface{})
	}

	return nil
}

func (c *Context) InitConfig(configFile string) error {
	v := config_viper.New()
	c.ConfigInterface = v
	return v.Init(configFile)
}

func (c *Context) InitLog(configPath string) error {

	l := logger_logrus.New()
	c.LoggerInterface = l

	return l.Init(c.Config(), "logger")
}

func (c *Context) InitDb(configPath string) error {

	d := db_gorm.New()

	c.db = d
	c.db.WithConfigBase = c.WithConfigBase
	c.db.WithLoggerBase = c.WithLoggerBase

	return d.Init(configPath)
}
