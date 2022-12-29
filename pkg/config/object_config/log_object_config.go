package object_config

import (
	"fmt"
	"reflect"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

func Info(log logger.Logger, obj Object, configPath ...string) {
	Log(log, utils.OptionalArg("", configPath...), obj)
}

func Debug(log logger.Logger, obj Object, configPath ...string) {
	Log(log, utils.OptionalArg("", configPath...), obj, logger.DebugLevel)
}

func Log(log logger.Logger, configPath string, obj Object, logLevel ...logger.Level) {

	level := utils.OptionalArg(logger.InfoLevel, logLevel...)

	cfg := reflect.ValueOf(obj.Config())
	typ := cfg.Type()
	if typ.Kind() == reflect.Ptr {
		typ = cfg.Elem().Type()
	}

	for i := 0; i < typ.NumField(); i++ {

		field := typ.Field(i)

		key := field.Name
		keyPath := Key(configPath, key)

		value := cfg.Elem().Field(i)
		formatValue := func() string {

			val := ""
			_, ok := field.Tag.Lookup("mask")
			if ok {
				val = "********"
			} else {
				val = fmt.Sprintf("%v", value)
			}

			return val
		}

		logParameter(log, keyPath, formatValue(), level)
	}
}

func logParameter(log logger.Logger, key string, value string, logLevel logger.Level) {
	log.Log(logLevel, "Configuration", logger.Fields{"key": key, "value": value})
}
