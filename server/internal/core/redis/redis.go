// Package redis provides a single shared *redis.Client for any subsystem that
// needs Redis (rate limiting, distributed locks, cache, etc.).
//
// asynq has its own internal Redis client and is not affected by this package
// — they just point at the same Redis server via the same env vars.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	return &Config{
		Addr:     viper.GetString("REDIS_ADDR"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}
}

// NewClient builds a connected *redis.Client. Boot does NOT fail if Redis is
// unreachable — callers (rate limiter, etc.) are expected to fail-open or
// degrade gracefully on Redis errors so a transient Redis outage doesn't take
// down dependent services entirely.
func NewClient(lc fx.Lifecycle, cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := client.Ping(pingCtx).Err(); err != nil {
				fmt.Printf("[redis] ping failed at startup: %v (callers should fail-open)\n", err)
			} else {
				fmt.Printf("[redis] connected to %s\n", cfg.Addr)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client
}

var Module = fx.Module("redis",
	fx.Provide(LoadConfig, NewClient),
)
