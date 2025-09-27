package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		c.Header("X-Correlation-ID", correlationID)

		ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
