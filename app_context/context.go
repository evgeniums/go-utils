package app_context

import (
	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/logger"
)

type Validator interface {
}

type Context interface {
	logger.WithLogger
	config.WithConfig

	DB() DB
	Validator() Validator

	Testing() bool
	TestParameters() map[string]interface{}
	SetTestParameter(key string, value interface{})
	GetTestParameter(key string) (interface{}, bool)
}
