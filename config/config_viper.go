package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ConfigViper struct {
	Viper *viper.Viper
}

func (c *ConfigViper) GetString(key string) string {
	return c.Viper.GetString(key)
}

func (c *ConfigViper) GetUint(key string) uint {
	return c.Viper.GetUint(key)
}

func (c *ConfigViper) Get(key string) interface{} {
	return c.Viper.Get(key)
}

func (c *ConfigViper) AllKeys() []string {
	return c.Viper.AllKeys()
}

func (c *ConfigViper) GetBool(key string) bool {
	return c.Viper.GetBool(key)
}

func (c *ConfigViper) Init(configFile string) error {

	c.Viper = viper.New()

	c.Viper.SetConfigFile(configFile)
	c.Viper.SetConfigType("json")
	c.Viper.AddConfigPath(".")
	err := c.Viper.ReadInConfig()
	if err != nil {
		msg := fmt.Errorf("fatal error config file: %s", err)
		return msg
	}

	includes := c.Viper.GetStringSlice("include")
	for _, include := range includes {
		c.Viper.SetConfigFile(include)
		err = c.Viper.MergeInConfig()
		if err != nil {
			msg := fmt.Errorf("failed to include config file %v: %s", include, err)
			return msg
		}
	}

	return nil
}
