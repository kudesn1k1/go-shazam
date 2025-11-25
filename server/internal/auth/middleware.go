package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
	authHeaderName   string     = "Authorization"
	bearerPrefix     string     = "Bearer "
)

// AuthMiddleware validates JWT tokens and adds user ID to context
func AuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authHeaderName)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			message := "invalid token"

			if err == ErrExpiredToken {
				message = "token has expired"
			}

			c.AbortWithStatusJSON(status, gin.H{"error": message})
			return
		}

		// Add user ID to context
		ctx := context.WithValue(c.Request.Context(), UserIDContextKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present, but doesn't require them
func OptionalAuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authHeaderName)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Add user ID to context
		ctx := context.WithValue(c.Request.Context(), UserIDContextKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}
