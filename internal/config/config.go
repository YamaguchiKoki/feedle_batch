package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
}

func Load() *Config {
	return &Config{
		SupabaseURL:        viper.GetString("SUPABASE_URL"),
		SupabaseServiceKey: viper.GetString("SUPABASE_SERVICE_KEY"),
	}
}
