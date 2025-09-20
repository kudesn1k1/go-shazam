package http

import "github.com/spf13/viper"

type Config struct {
	Port int
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	return &Config{
		Port: viper.GetInt("APP_PORT"),
	}
}
