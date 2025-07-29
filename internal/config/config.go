package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
	RedditClientID     string
	RedditClientSecret string
	RedditUsername     string
}

func Load() *Config {
	return &Config{
		SupabaseURL:        viper.GetString("SUPABASE_URL"),
		SupabaseServiceKey: viper.GetString("SUPABASE_SERVICE_KEY"),
		RedditClientID:     viper.GetString("REDDIT_CLIENT_ID"),
		RedditClientSecret: viper.GetString("REDDIT_CLIENT_SECRET"),
		RedditUsername:     viper.GetString("REDDIT_USERNAME"),
	}
}
