package config_viper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/spf13/viper"
	"github.com/tidwall/jsonc"
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

func ReadConfigFromFile(v *viper.Viper, path string, cfgType string) error {

	v.SetConfigType(cfgType)

	if cfgType != "json" {
		v.SetConfigFile(path)
		err := v.ReadInConfig()
		if err != nil {
			return fmt.Errorf("fatal error while reading config file: %s", err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("fatal error while reading config file: %s", err)
		}
		b := jsonc.ToJSON(data)
		if err = v.ReadConfig(bytes.NewReader(b)); err != nil {
			msg := fmt.Errorf("failed to parse configuration JSON: %s", err)
			return msg
		}
	}
	return nil
}

func (c *ConfigViper) LoadFile(configFile string, configType ...string) error {

	// setup
	cfgType := utils.OptionalArg("json", configType...)
	c.Viper.SetConfigFile(configFile)
	c.Viper.SetConfigType(cfgType)

	// read main configuration file
	err := ReadConfigFromFile(c.Viper, configFile, cfgType)
	if err != nil {
		return fmt.Errorf("fatal error while reading config file: %s", err)
	}

	// load includes
	includes := c.Viper.GetStringSlice("include")
	for _, include := range includes {
		if !utils.FileExists(include) {
			// try relative path
			newInclude := filepath.Join(filepath.Dir(configFile), include)
			if !utils.FileExists(newInclude) {
				return fmt.Errorf("failed to include config file %s or %s", include, newInclude)
			}
			include = newInclude
		}
		c.Viper.SetConfigFile(include)
		err = c.Viper.MergeInConfig()
		if err != nil {
			return fmt.Errorf("failed to include config file %s: %s", include, err)
		}
	}

	// load includes for customizable merging
	mergeSectionKey := "include_merge"
	if c.Viper.IsSet(mergeSectionKey) {

		mergeSection := c.Viper.Get(mergeSectionKey)
		includeItems, ok := mergeSection.([]interface{})
		if !ok {
			return fmt.Errorf("invalid format of %s section", mergeSectionKey)
		}
		for i := range includeItems {
			item := &IncludeMerge{}
			itemKey := fmt.Sprintf("%s.%d", mergeSectionKey, i)
			err = c.Viper.UnmarshalKey(itemKey, item)
			if err != nil {
				return fmt.Errorf("invalid format of %s", itemKey)
			}

			if !utils.FileExists(item.Path) {
				// try relative path
				newPath := filepath.Join(filepath.Dir(configFile), item.Path)
				if !utils.FileExists(newPath) {
					return fmt.Errorf("failed to include config file %s or %s", item.Path, newPath)
				}
				item.Path = newPath
			}

			cfg := viper.New()
			err = ReadConfigFromFile(cfg, item.Path, cfgType)
			if err != nil {
				return fmt.Errorf("failed to read configuration from %s: %s", item.Path, err)
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
							return fmt.Errorf("failed to append array from %s: %s is not array in main config", item.Path, to)
						}
						newSlice, ok := newData.([]interface{})
						if !ok {
							return fmt.Errorf("failed to append array from %s: %s is not array in included config", item.Path, from)
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
			return fmt.Errorf("failed to marshal viper settings: %s", err)
		}
		c.Viper = viper.New()
		c.Viper.SetConfigType(cfgType)
		if err = c.Viper.ReadConfig(bytes.NewReader(b)); err != nil {
			return fmt.Errorf("failed to re-read viper settings: %s", err)
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
