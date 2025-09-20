package db

import (
	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func LoadDBConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	return &Config{
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetInt("DB_PORT"),
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		Database: viper.GetString("DB_DATABASE"),
	}
}
