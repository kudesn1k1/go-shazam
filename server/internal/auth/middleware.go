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
	RolesContextKey  contextKey = "roles"
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

		ctx := context.WithValue(c.Request.Context(), UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, RolesContextKey, claims.Roles)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}

func GetRolesFromContext(ctx context.Context) []string {
	roles, _ := ctx.Value(RolesContextKey).([]string)
	return roles
}

// RoleMiddleware checks that the authenticated user has at least one of the required roles
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetRolesFromContext(c.Request.Context())
		for _, userRole := range roles {
			for _, required := range requiredRoles {
				if userRole == required {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
