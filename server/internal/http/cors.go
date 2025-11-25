package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type CORSConfig struct {
	AllowedOrigins []string
}

func LoadCORSConfig() *CORSConfig {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")

	origins := viper.GetStringSlice("CORS_ALLOWED_ORIGINS")
	if len(origins) == 0 {
		originsStr := viper.GetString("CORS_ALLOWED_ORIGINS")
		if originsStr != "" {
			origins = []string{originsStr}
		}
	}

	return &CORSConfig{
		AllowedOrigins: origins,
	}
}

func SetupCORS(r *gin.Engine, config *CORSConfig) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Correlation-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Correlation-ID"},
		AllowCredentials: true, // Required for cookies
		MaxAge:           12 * time.Hour,
	}))
}
