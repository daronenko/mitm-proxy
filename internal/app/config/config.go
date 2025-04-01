package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to app config file")
}

type Config struct {
	App Spec `mapstructure:"app"`
}

func New() (*Config, error) {
	flag.Parse()

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for _, key := range viper.AllKeys() {
		value := viper.Get(key)
		viper.Set(key, value)
	}

	conf := &Config{}
	if err := viper.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return conf, nil
}
