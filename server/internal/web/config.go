package web

import "github.com/spf13/viper"

type Config struct {
	// PublicOrigin is the URL prefix the site is served from (e.g.
	// "https://goshazam.example"). Used to build absolute URLs in canonical
	// tags, OG tags, and the sitemap. Defaults to empty (relative URLs only)
	// if not set, which still works but loses some SEO benefit.
	PublicOrigin string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	return &Config{
		PublicOrigin: viper.GetString("PUBLIC_ORIGIN"),
	}
}
