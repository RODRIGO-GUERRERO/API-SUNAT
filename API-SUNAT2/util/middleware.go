package util

import (
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(`{"timestamp":"%s","status":%d,"latency":"%s","client_ip":"%s","method":"%s","path":"%s","user_agent":"%s","error_message":"%s"}%s`,
			param.TimeStamp.Format(time.RFC3339),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
			"\n",
		)
	})
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateCorrelationID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

func generateCorrelationID() string {
	return uuid.New().String()
}

// GenerateCorrelationID es la versión pública de generateCorrelationID
func GenerateCorrelationID() string {
	return uuid.New().String()
} 