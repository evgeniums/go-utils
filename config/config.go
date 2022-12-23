package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/validator"
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
}

type WithConfigBase struct {
	cfg Config
}

func (w *WithConfigBase) Config() Config {
	return w.cfg
}

func (w *WithConfigBase) SetConfig(cfg Config) {
	w.cfg = cfg
}

func Key(path string, key string) string {
	if path == "" {
		return key
	}
	return fmt.Sprintf("%v.%v", path, key)
}

func Path(defaultPath string, optional ...string) string {
	path := defaultPath
	if len(optional) > 0 {
		path = optional[0]
	}
	return path
}

func logParameter(log logger.Logger, key string, value interface{}) {
	log.Info("Configuration", logger.Fields{"key": key, "value": value})
}

func LogConfigParameters(cfg Config, log logger.Logger, configPaths ...string) {
	configPath := ""
	if len(configPaths) > 0 {
		configPath = configPaths[0]
	}
	keys := cfg.AllKeys()
	for _, key := range keys {
		k := strings.TrimPrefix(key, configPath)
		k = strings.TrimPrefix(k, ".")
		if !strings.Contains(k, ".") && (configPath == "" || strings.HasPrefix(key, configPath)) {
			value := cfg.Get(key)
			if strings.Contains(key, "password") || strings.Contains(key, "passphrase") || strings.Contains(key, "secret") {
				value = "*******"
			}
			logParameter(log, key, value)
		}
	}
}

type logHandler = func(cfg Config, path string)

func initObject(cfg Config, log logHandler, vld validator.Validator, object interface{}, defaultConfigPath string, optionalConfigPath ...string) error {
	path := Path(defaultConfigPath, optionalConfigPath...)
	_, err := LoadConfiguration(cfg, path, object)
	if err != nil {
		return err
	}
	log(cfg, path)
	err = vld.Validate(object)
	if err != nil {
		return err
	}
	return nil
}

func InitObject(cfg Config, log logger.Logger, vld validator.Validator, object interface{}, defaultConfigPath string, optionalConfigPath ...string) error {
	logHandler := func(cfg Config, path string) {
		LogConfigParameters(cfg, log, path)
	}
	return initObject(cfg, logHandler, vld, object, defaultConfigPath, optionalConfigPath...)
}

func InitObjectNoLog(cfg Config, vld validator.Validator, object interface{}, defaultConfigPath string, optionalConfigPath ...string) error {
	logHandler := func(cfg Config, path string) {}
	return initObject(cfg, logHandler, vld, object, defaultConfigPath, optionalConfigPath...)
}

func LoadConfiguration(cfg Config, configPath string, object interface{}) (skippedKeys []string, err error) {
	// Get a reflect.Value for the object.
	objectValue := reflect.ValueOf(object)

	// Make sure the object is a pointer to a struct.
	if objectValue.Kind() != reflect.Ptr || objectValue.Elem().Kind() != reflect.Struct {
		return nil, errors.New("object must be a pointer to a struct")
	}

	// Get a reflect.Type for the struct.
	objectType := objectValue.Elem().Type()

	// Iterate over the fields of the struct.
	for i := 0; i < objectType.NumField(); i++ {
		// Get the field and its tags.
		field := objectType.Field(i)
		configTag, ok := field.Tag.Lookup("config")
		if !ok {
			// Skip fields without a "config" tag.
			continue
		}

		// Construct the full config path for this field.
		fieldConfigPath := Key(configPath, configTag)

		// Get the field value.
		fieldValue := objectValue.Elem().Field(i)

		// Set the default value if the field is zero-valued.
		if fieldValue.IsZero() {
			defaultTag, ok := field.Tag.Lookup("default")
			if ok {
				cfg.SetDefault(fieldConfigPath, defaultTag)
			}
		}

		switch field.Type.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			fieldValue.SetInt(int64(cfg.GetInt(fieldConfigPath)))
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			fieldValue.SetUint(uint64(cfg.GetUint(fieldConfigPath)))
		case reflect.Float64, reflect.Float32:
			fieldValue.SetFloat(cfg.GetFloat64(fieldConfigPath))
		case reflect.String:
			fieldValue.SetString(cfg.GetString(fieldConfigPath))
		case reflect.Bool:
			fieldValue.SetBool(cfg.GetBool(fieldConfigPath))
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Int {
				slice := cfg.GetIntSlice(fieldConfigPath)
				fieldValue.Set(reflect.ValueOf(slice))
			} else {
				slice := cfg.GetStringSlice(fieldConfigPath)
				fieldValue.Set(reflect.ValueOf(slice))
			}
		default:
			skippedKeys = append(skippedKeys, configTag)
		}
	}

	return skippedKeys, nil
}
