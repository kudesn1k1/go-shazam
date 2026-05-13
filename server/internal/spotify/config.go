package spotify

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ClientID       string
	ClientSecret   string
	RequestTimeout time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
	RateLimitRPS   float64
	RateLimitBurst int
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	if viper.GetString("SPOTIFY_CLIENT_ID") == "" || viper.GetString("SPOTIFY_CLIENT_SECRET") == "" {
		panic("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be set")
	}

	viper.SetDefault("SPOTIFY_REQUEST_TIMEOUT", "10s")
	viper.SetDefault("SPOTIFY_MAX_RETRIES", 3)
	viper.SetDefault("SPOTIFY_RETRY_BACKOFF", "500ms")
	viper.SetDefault("SPOTIFY_RATE_LIMIT_RPS", 10.0)
	viper.SetDefault("SPOTIFY_RATE_LIMIT_BURST", 20)

	return &Config{
		ClientID:       viper.GetString("SPOTIFY_CLIENT_ID"),
		ClientSecret:   viper.GetString("SPOTIFY_CLIENT_SECRET"),
		RequestTimeout: viper.GetDuration("SPOTIFY_REQUEST_TIMEOUT"),
		MaxRetries:     viper.GetInt("SPOTIFY_MAX_RETRIES"),
		RetryBackoff:   viper.GetDuration("SPOTIFY_RETRY_BACKOFF"),
		RateLimitRPS:   viper.GetFloat64("SPOTIFY_RATE_LIMIT_RPS"),
		RateLimitBurst: viper.GetInt("SPOTIFY_RATE_LIMIT_BURST"),
	}
}
