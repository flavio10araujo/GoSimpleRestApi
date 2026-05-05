package config

import (
	"github.com/spf13/viper"
)

type PaginationConfig struct {
	DefaultLimit int
	MaxLimit     int
}

func LoadPaginationConfig() *PaginationConfig {
	viper.SetDefault("pagination.default_limit", 20)
	viper.SetDefault("pagination.max_limit", 100)

	viper.BindEnv("pagination.default_limit", "PAGINATION_DEFAULT_LIMIT")
	viper.BindEnv("pagination.max_limit", "PAGINATION_MAX_LIMIT")

	return &PaginationConfig{
		DefaultLimit: viper.GetInt("pagination.default_limit"),
		MaxLimit:     viper.GetInt("pagination.max_limit"),
	}
}
