package spotify

import (
	"github.com/spf13/viper"
)

type Config struct {
	ClientID     string
	ClientSecret string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	if viper.GetString("SPOTIFY_CLIENT_ID") == "" || viper.GetString("SPOTIFY_CLIENT_SECRET") == "" {
		panic("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be set")
	}

	return &Config{
		ClientID:     viper.GetString("SPOTIFY_CLIENT_ID"),
		ClientSecret: viper.GetString("SPOTIFY_CLIENT_SECRET"),
	}
}
