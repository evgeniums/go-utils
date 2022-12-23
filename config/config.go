package config

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/logger"
)

type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	SetDefault(key string, value interface{})

	AllKeys() []string
}

type WithConfig interface {
	Config() Config
	LogConfigParameters(log logger.Logger, configPaths ...string)
}

type WithConfigBase struct {
	ConfigInterface Config
}

func (w *WithConfigBase) Config() Config {
	return w.ConfigInterface
}

func Key(path string, key string) string {
	if path == "" {
		return key
	}
	return fmt.Sprintf("%v.%v", path, key)
}

func LogParameter(log logger.Logger, key string, value interface{}) {
	log.Info("Configuration", logger.Fields{"key": key, "value": value})
}

func (w *WithConfigBase) LogConfigParameters(log logger.Logger, configPaths ...string) {
	configPath := ""
	if len(configPaths) > 0 {
		configPath = configPaths[0]
	}
	keys := w.ConfigInterface.AllKeys()
	for _, key := range keys {
		k := strings.TrimPrefix(key, configPath)
		k = strings.TrimPrefix(k, ".")
		if !strings.Contains(k, ".") && (configPath == "" || strings.HasPrefix(key, configPath)) {
			value := w.ConfigInterface.Get(key)
			if strings.Contains(key, "password") || strings.Contains(key, "passphrase") || strings.Contains(key, "secret") {
				value = "*******"
			}
			LogParameter(log, key, value)
		}
	}
}
