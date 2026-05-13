package http

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewGinRouter, NewHttpServer, LoadConfig, LoadCORSConfig),
)

func NewGinRouter(lc fx.Lifecycle, corsConfig *CORSConfig) *gin.Engine {
	r := gin.Default()

	// Setup CORS before other middleware
	SetupCORS(r, corsConfig)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	SetupSecurityHeader(r)

	return r
}

func SetupSecurityHeader(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data: blob:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		// microphone=(self) lets the SPA's recognition feature call getUserMedia
		// on same-origin pages while still blocking any cross-origin iframe from
		// piggy-backing on the permission. The other features stay denied because
		// nothing in the app uses them — flip to (self) only when needed.
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(self),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Next()
	})
}

func NewHttpServer(lc fx.Lifecycle, r *gin.Engine, config *Config) *http.Server {
	srv := &http.Server{Addr: ":" + fmt.Sprint(config.Port), Handler: r}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			fmt.Println("Server is listening on", srv.Addr)

			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
