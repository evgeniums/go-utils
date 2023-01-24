package config_viper

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/spf13/viper"
)

type ConfigViper struct {
	*viper.Viper
}

func New() *ConfigViper {
	v := &ConfigViper{}
	v.Viper = viper.New()
	return v
}

func (c *ConfigViper) LoadFile(configFile string, configType ...string) error {

	cfgType := utils.OptionalArg("json", configType...)
	c.Viper.SetConfigFile(configFile)
	c.Viper.SetConfigType(cfgType)
	c.Viper.AddConfigPath(".")
	err := c.Viper.ReadInConfig()
	if err != nil {
		msg := fmt.Errorf("fatal error while reading config file: %s", err)
		return msg
	}

	includes := c.Viper.GetStringSlice("include")
	for _, include := range includes {
		c.Viper.SetConfigFile(include)
		err = c.Viper.MergeInConfig()
		if err != nil {
			msg := fmt.Errorf("failed to include config file %s: %s", include, err)
			return msg
		}
	}

	return nil
}

func (c *ConfigViper) LoadString(configStr string, configType ...string) error {

	cfgType := utils.OptionalArg("json", configType...)
	c.Viper.SetConfigType(cfgType)

	err := c.Viper.ReadConfig(strings.NewReader(configStr))
	if err != nil {
		msg := fmt.Errorf("fatal error while reading config string: %s", err)
		return msg
	}

	return nil
}
