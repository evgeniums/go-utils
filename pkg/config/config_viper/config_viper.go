package config_viper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
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

	// setup
	cfgType := utils.OptionalArg("json", configType...)
	c.Viper.SetConfigFile(configFile)
	c.Viper.SetConfigType(cfgType)
	c.Viper.AddConfigPath(".")

	// read main configuration file
	err := c.Viper.ReadInConfig()
	if err != nil {
		msg := fmt.Errorf("fatal error while reading config file: %s", err)
		return msg
	}

	// load includes
	includes := c.Viper.GetStringSlice("include")
	for _, include := range includes {
		if !utils.FileExists(include) {
			// try relative path
			newInclude := filepath.Join(filepath.Dir(configFile), include)
			if !utils.FileExists(newInclude) {
				err = fmt.Errorf("failed to include config file %s or %s", include, newInclude)
				return err
			}
			include = newInclude
		}
		c.Viper.SetConfigFile(include)
		err = c.Viper.MergeInConfig()
		if err != nil {
			msg := fmt.Errorf("failed to include config file %s: %s", include, err)
			return msg
		}
	}

	// load includes with arrays for merging
	if c.Viper.IsSet("include_arrays") {

		// load and merge arrays data
		includeArrays := c.Viper.Get("include_arrays")
		includeArraysSlice, ok := includeArrays.([]interface{})
		if !ok {
			msg := fmt.Errorf("invalid format of include_arrays section")
			return msg
		}
		for i := range includeArraysSlice {
			includes := c.Viper.GetStringMapStringSlice(fmt.Sprintf("include_arrays.%d", i))
			for include, keys := range includes {
				if !utils.FileExists(include) {
					// try relative path
					newInclude := filepath.Join(filepath.Dir(configFile), include)
					if !utils.FileExists(newInclude) {
						err = fmt.Errorf("failed to include array config file %s or %s", include, newInclude)
						return err
					}
					include = newInclude
				}
				arrCfg := viper.New()
				arrCfg.SetConfigFile(include)
				arrCfg.SetConfigType(cfgType)
				err = arrCfg.ReadInConfig()
				if err != nil {
					msg := fmt.Errorf("failed to include array config file %s: %s", include, err)
					return msg
				}

				for _, key := range keys {
					if arrCfg.IsSet(key) {
						if !c.Viper.IsSet(key) {
							c.Viper.Set(key, arrCfg.Get(key))
						} else {
							arr1 := c.Viper.Get(key)
							arr2 := arrCfg.Get(key)
							arr1Slice, ok := arr1.([]interface{})
							if !ok {
								msg := fmt.Errorf("failed to include array config file %s: %s is not array in main config", include, key)
								return msg
							}
							arr2Slice, ok := arr2.([]interface{})
							if !ok {
								msg := fmt.Errorf("failed to include array config file %s: %s is not array in included config", include, key)
								return msg
							}
							merge := append(arr1Slice, arr2Slice...)
							c.Viper.Set(key, merge)
						}
					}
				}
			}
		}

		// reload viper configuration
		all := c.Viper.AllSettings()
		b, err := json.MarshalIndent(all, "", " ")
		if err != nil {
			msg := fmt.Errorf("failed to marshal viper settings: %s", err)
			return msg
		}
		c.Viper = viper.New()
		c.Viper.SetConfigType(cfgType)
		if err = c.Viper.ReadConfig(bytes.NewReader(b)); err != nil {
			msg := fmt.Errorf("failed to re-read viper settings: %s", err)
			return msg
		}
	}

	// done
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
