package app_context

import (
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type BuildConfig struct {
	Version  string
	Time     string
	Revision string
}

type Context interface {
	logger.WithLogger
	config.WithCfg
	db.WithDB

	Cache() cache.Cache
	Validator() validator.Validator

	Testing() bool
	TestParameters() map[string]interface{}
	SetTestParameter(key string, value interface{})
	GetTestParameter(key string) (interface{}, bool)

	AppInstance() string
	Application() string
	Hostname() string

	SetPublisher(publisher pubsub.Publisher)
	Publisher() pubsub.Publisher

	Close()
}
