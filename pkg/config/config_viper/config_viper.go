package config_viper

import (
	"fmt"

	"github.com/spf13/viper"
)

type ConfigViper struct {
	viper.Viper
}

func New() *ConfigViper {
	return &ConfigViper{}
}

func (c *ConfigViper) Init(configFile string) error {

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
