package config

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/logger"
)

type Config interface {
	Init(configFile string) error

	GetString(key string) string
	GetUint(key string) uint
	GetBool(key string) bool
	Get(key string) interface{}

	AllKeys() []string
}

type WithConfig interface {
	Config() Config
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
