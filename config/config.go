package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	RedisAddress string `mapstructure:"REDISADDRESS"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}
	var conf *Config
	err = v.Unmarshal(conf, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot write config to struct: %w", err)
	}

	return conf, nil
}
