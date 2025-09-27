package youtube

import "github.com/spf13/viper"

type Config struct {
	ApiKey string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	return &Config{
		ApiKey: viper.GetString("YOUTUBE_API_KEY"),
	}
}
