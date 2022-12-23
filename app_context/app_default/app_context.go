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
	testing    bool `config:"testing"`
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
	err := c.initConfig(configFile)
	if err != nil {
		return err
	}

	// setup logger
	logConfigPath := "log"
	c.initLog(logConfigPath)
	log := c.Logger()
	log.Info("Starting...")

	// log build version
	log.Info("Build configuration", logger.Fields{"build_time": Time, "package_version": Version, "git_revision": Revision})

	// log logger configuration
	config.LogConfigParameters(c.Config(), log, logConfigPath)

	// load top level configuration
	err = config.InitObject(c.Config(), log, c.validator, c, "")
	if err != nil {
		return log.Fatal("failed to load application configuration", err)
	}

	// connect to DB
	err = c.initDb()
	if err != nil {
		return err
	}

	// setup testing
	if c.Testing() {
		log.Info("Running in test mode")
		c.testValues = make(map[string]interface{})
	}

	return nil
}

func (c *Context) initConfig(configFile string) error {
	v := config_viper.New()
	c.SetConfig(v)
	return v.Init(configFile)
}

func (c *Context) initLog(configPath string) error {
	l := logger_logrus.New()
	c.SetLogger(l)
	return l.Init(c.Config(), c.validator, "logger")
}

func (c *Context) initDb() error {
	d := db_gorm.New(c.Logger())
	c.db = d
	return d.Init(c.Config(), c.validator, "psql")
}
