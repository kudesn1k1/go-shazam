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
	fx.Provide(NewGinRouter, NewHttpServer, LoadConfig),
	fx.Invoke(func(r *http.Server) {}),
)

func NewGinRouter(lc fx.Lifecycle) *gin.Engine {
	r := gin.Default()

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
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
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
