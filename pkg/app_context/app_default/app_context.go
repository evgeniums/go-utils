package app_default

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/cache/inmem_cache"
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
	TESTING      bool
	APP_INSTANCE string
	HOSTNAME     string

	APPLICATION string
}

type WithInitGormDb interface {
	InitDB(configPath string, gormDbConnector ...*db_gorm.DbConnector) error
}

type Context struct {
	logger.WithLoggerBase
	config.WithCfgBase

	db           *db_gorm.GormDB
	validator    *validator_playground.PlaygroundValdator
	cache        cache.Cache
	inmemCache   *inmem_cache.InmemCache[string]
	logrusLogger *logger_logrus.LogrusLogger

	contextConfig

	testValues  map[string]interface{}
	initialized bool
}

func (c *Context) Config() interface{} {
	return &c.contextConfig
}

func (c *Context) Db() db.DB {
	return c.db
}

func (c *Context) Cache() cache.Cache {
	return c.cache
}

func (c *Context) Validator() validator.Validator {
	return c.validator
}

func (c *Context) Testing() bool {
	return c.TESTING
}

func (c *Context) AppInstance() string {
	return c.APP_INSTANCE
}

func Application() string {
	proc, _ := os.Executable()
	return filepath.Base(proc)
}

func (c *Context) Application() string {
	if c.APPLICATION != "" {
		return c.APPLICATION
	}
	return Application()
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

type AppConfig struct {
	Cache cache.Cache
}

func (a *AppConfig) GetCache() cache.Cache {
	return a.Cache
}

type AppConfigI interface {
	GetCache() cache.Cache
}

func New(buildConfig *app_context.BuildConfig, appConfig ...AppConfigI) *Context {

	if buildConfig != nil {
		Version = buildConfig.Version
		Time = buildConfig.Time
		Revision = buildConfig.Revision
	}

	rand.Seed(time.Now().UnixNano())

	c := &Context{}
	c.validator = validator_playground.New()

	if len(appConfig) != 0 {
		c.cache = appConfig[0].GetCache()
	}

	if c.cache == nil {
		c.inmemCache = inmem_cache.New[string]()
		c.cache = cache.New(c.inmemCache)
		c.inmemCache.Start()
	}

	c.logrusLogger = logger_logrus.New()
	c.WithLoggerBase.Init(c.logrusLogger)

	return c
}

func (c *Context) InitWithArgs(configFile string, args []string, configType ...string) error {

	if c.initialized {
		return nil
	}

	app_context.SetTimeZone()

	// load configuration
	fmt.Printf("Application %s using configuration file %s\n", c.Application(), configFile)
	err := c.initConfig(configFile)
	if err != nil {
		return c.Logger().PushFatalStack("failed to load application configuration", err)
	}

	// load command line arguments
	err = config.LoadArgs(c.Cfg(), args)
	if err != nil {
		return c.Logger().PushFatalStack("failed to override confiuration parameters", err)
	}

	// setup logger
	logConfigPath := "logger"
	err = c.initLog(logConfigPath)
	if err != nil {
		return c.Logger().PushFatalStack("failed to init application logger", err)
	}
	log := c.Logger()

	// log build version
	log.Info("Build configuration", logger.Fields{"build_time": Time, "package_version": Version, "git_revision": Revision})
	fmt.Printf("Build configuration: build_time=%s, package_version=%s, get_revision=%s\n", Time, Version, Revision)

	// log logger configuration
	object_config.Info(log, c.logrusLogger, logConfigPath)

	// load top level application configuration
	err = object_config.LoadLogValidate(c.Cfg(), log, c.validator, c, "")
	if err != nil {
		return log.PushFatalStack("failed to init application configuration", err)
	}

	// setup testing
	if c.Testing() {
		log.Info("Running in test mode")
		c.testValues = make(map[string]interface{})
	}

	// done
	c.initialized = true
	return nil
}

func (c *Context) Init(configFile string, configType ...string) error {
	return c.InitWithArgs(configFile, nil, configType...)
}

func (c *Context) Close() {
	if c.db != nil {
		c.db.Close()
	}
	if c.inmemCache != nil {
		c.inmemCache.Stop()
	}
}

func (c *Context) initConfig(configFile string, configType ...string) error {

	if configFile == "" {
		return errors.New("configuration file not specified")
	}

	v := config_viper.New()
	c.SetCfg(v)
	err := v.LoadFile(configFile, configType...)
	if err != nil {
		return err
	}
	object_config.Load(v, c, "")
	return nil
}

func (c *Context) initLog(configPath string) error {
	return c.logrusLogger.Init(c.Cfg(), c.validator, configPath)
}

func (c *Context) InitDB(configPath string, gormDbConnector ...*db_gorm.DbConnector) error {
	if c.db != nil {
		return nil
	}
	d := db_gorm.New(gormDbConnector...)
	c.db = d
	return d.Init(c, c.Cfg(), c.validator, configPath)
}

func (c *Context) Hostname() string {
	if c.HOSTNAME != "" {
		return c.HOSTNAME
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}
	return hostname
}
