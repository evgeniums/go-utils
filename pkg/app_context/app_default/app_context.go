package app_default

import (
	"math/rand"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/config_viper"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/logger/logger_logrus"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"github.com/evgeniums/go-backend-helpers/pkg/validator/validator_playground"
)

var Version = "development"
var Time = "unknown"
var Revision = "unknown"

type contextConfig struct {
	TESTING bool
}

type Context struct {
	logger.WithLoggerBase
	config.WithCfgBase

	db        *db_gorm.GormDB
	validator *validator_playground.PlaygroundValdator

	contextConfig

	testValues map[string]interface{}
}

func (c *Context) Config() interface{} {
	return &c.contextConfig
}

func (c *Context) DB() db.DB {
	return c.db
}

func (c *Context) Validator() validator.Validator {
	return c.validator
}

func (c *Context) Testing() bool {
	return c.TESTING
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

func (c *Context) Init(configFile string, configType ...string) error {

	// load configuration
	err := c.initConfig(configFile)
	if err != nil {
		return err
	}

	// setup logger
	logConfigPath := "log"
	l, err := c.initLog(logConfigPath)
	if err != nil {
		return err
	}
	log := c.Logger()
	log.Info("Starting...")

	// log build version
	log.Info("Build configuration", logger.Fields{"build_time": Time, "package_version": Version, "git_revision": Revision})

	// log logger configuration
	object_config.Info(log, l, logConfigPath)

	// load top level application configuration
	err = object_config.LoadLogValidate(c.Cfg(), log, c.validator, c, "")
	if err != nil {
		return log.Fatal("failed to load application configuration", err)
	}

	// init main application database
	err = c.initDb()
	if err != nil {
		return err
	}

	// setup testing
	if c.Testing() {
		log.Info("Running in test mode")
		c.testValues = make(map[string]interface{})
	}

	// done
	return nil
}

func (c *Context) initConfig(configFile string, configType ...string) error {
	v := config_viper.New()
	c.SetCfg(v)
	return v.LoadFile(configFile, configType...)
}

func (c *Context) initLog(configPath string) (*logger_logrus.LogrusLogger, error) {
	l := logger_logrus.New()
	c.WithLoggerBase.Init(l)
	return l, l.Init(c.Cfg(), c.validator, configPath)
}

func (c *Context) initDb() error {
	d := db_gorm.New()
	c.db = d
	return d.Init(c, c.Cfg(), c.validator, "psql")
}
