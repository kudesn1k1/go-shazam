package auth

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	EmailEncryptionKey string
	CookieDomain       string
	CookieSecure       bool // true for HTTPS
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("JWT_ACCESS_SECRET", "change-me-access-secret-key-32ch")
	viper.SetDefault("JWT_REFRESH_SECRET", "change-me-refresh-secret-key-32c")
	viper.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_TTL_DAYS", 7)
	viper.SetDefault("EMAIL_ENCRYPTION_KEY", "change-me-email-encryption-key!")
	viper.SetDefault("COOKIE_DOMAIN", "")
	viper.SetDefault("COOKIE_SECURE", false)

	return &Config{
		AccessTokenSecret:  viper.GetString("JWT_ACCESS_SECRET"),
		RefreshTokenSecret: viper.GetString("JWT_REFRESH_SECRET"),
		AccessTokenTTL:     time.Duration(viper.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
		RefreshTokenTTL:    time.Duration(viper.GetInt("JWT_REFRESH_TTL_DAYS")) * 24 * time.Hour,
		EmailEncryptionKey: viper.GetString("EMAIL_ENCRYPTION_KEY"),
		CookieDomain:       viper.GetString("COOKIE_DOMAIN"),
		CookieSecure:       viper.GetBool("COOKIE_SECURE"),
	}
}
