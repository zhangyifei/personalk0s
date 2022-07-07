package cmdconfig

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type ConfigLoader struct {
	Paths []string
}

const DEFAULT_CONFIG = "defaultconfig/eke.cmd.default.yaml"

//go:embed defaultconfig/eke.cmd.default.yaml
var df embed.FS

func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		Paths: configPaths,
	}
}

func (c *ConfigLoader) Load() (*EkeCmdConfig, error) {
	v, err := c.preload()
	if err != nil {
		return nil, err
	}

	var ekeCmdConfig EkeCmdConfig
	if err := v.Unmarshal(&ekeCmdConfig); err != nil {
		return nil, err
	}

	return &ekeCmdConfig, nil
}

// Load reads the configuration files from disks and merges them
func (c *ConfigLoader) preload() (*viper.Viper, error) {
	v, _ := c.setDefault()

	if len(c.Paths) == 0 {
		return v, nil
	}

	for _, path := range c.Paths {
		if err := mergeConfig(v, path); err != nil {
			return viper.New(), err
		}
	}
	return v, nil
}

func (c *ConfigLoader) setDefault() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	data, _ := df.ReadFile(DEFAULT_CONFIG)

	v.ReadConfig(bytes.NewBuffer(data))

	return v, nil
}

func mergeConfig(v *viper.Viper, extraConfigPath string) error {
	cfgFile := filepath.Join(extraConfigPath, "eke.cmd.yaml")

	_, err := os.Stat(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	v.SetConfigFile(cfgFile)

	return v.MergeInConfig()
}
