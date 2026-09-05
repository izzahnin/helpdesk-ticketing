package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-Id", requestID)
		c.Header("X-Request-Id", requestID)

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		log.Printf(`{"request_id":"%s","method":"%s","path":"%s","status":%d,"latency_ms":%d}`,
			requestID, c.Request.Method, path, c.Writer.Status(), time.Since(start).Milliseconds())
	}
}
