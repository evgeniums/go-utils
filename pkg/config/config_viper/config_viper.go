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

const (
	MergeValue string = "value"
	MergeArray string = "array"
)

type IncludeMerge struct {
	Path string            `json:"path"`
	Mode string            `json:"path1"`
	Map  map[string]string `json:"map"`
}

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

	// load includes for customizable merging
	mergeSectionKey := "include_merge"
	if c.Viper.IsSet(mergeSectionKey) {

		mergeSection := c.Viper.Get(mergeSectionKey)
		includeItems, ok := mergeSection.([]interface{})
		if !ok {
			msg := fmt.Errorf("invalid format of %s section", mergeSectionKey)
			return msg
		}
		for i := range includeItems {
			item := &IncludeMerge{}
			itemKey := fmt.Sprintf("%s.%d", mergeSectionKey, i)
			err = c.Viper.UnmarshalKey(itemKey, item)
			if err != nil {
				msg := fmt.Errorf("invalid format of %s", itemKey)
				return msg
			}

			if !utils.FileExists(item.Path) {
				// try relative path
				newPath := filepath.Join(filepath.Dir(configFile), item.Path)
				if !utils.FileExists(newPath) {
					err = fmt.Errorf("failed to include config file %s or %s", item.Path, newPath)
					return err
				}
				item.Path = newPath
			}

			cfg := viper.New()
			cfg.SetConfigFile(item.Path)
			cfg.SetConfigType(cfgType)
			err = cfg.ReadInConfig()
			if err != nil {
				msg := fmt.Errorf("failed to read configuration from %s: %s", item.Path, err)
				return msg
			}

			for from, to := range item.Map {
				if to == "" {
					to = from
				}

				if cfg.IsSet(from) {
					if !c.Viper.IsSet(to) || item.Mode != MergeArray {
						c.Viper.Set(to, cfg.Get(from))
					} else {
						oldData := c.Viper.Get(to)
						newData := cfg.Get(from)

						oldSlice, ok := oldData.([]interface{})
						if !ok {
							msg := fmt.Errorf("failed to append array from %s: %s is not array in main config", item.Path, to)
							return msg
						}
						newSlice, ok := newData.([]interface{})
						if !ok {
							msg := fmt.Errorf("failed to append array from %s: %s is not array in included config", item.Path, from)
							return msg
						}
						merge := append(oldSlice, newSlice...)
						c.Viper.Set(to, merge)
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
