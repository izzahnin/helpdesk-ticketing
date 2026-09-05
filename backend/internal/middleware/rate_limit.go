package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	Mu        sync.Mutex
	Count     int
	ResetTime time.Time
}

var buckets sync.Map

func ResetRateLimitBuckets() {
	buckets.Range(func(key, _ interface{}) bool {
		buckets.Delete(key)
		return true
	})
}

func RateLimit(name string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := name + ":" + c.ClientIP()
		now := time.Now()
		value, _ := buckets.LoadOrStore(key, &bucket{Count: 0, ResetTime: now.Add(window)})
		b := value.(*bucket)
		b.Mu.Lock()
		if now.After(b.ResetTime) {
			b.Count = 0
			b.ResetTime = now.Add(window)
		}
		b.Count++
		if b.Count > limit {
			b.Mu.Unlock()
			c.JSON(429, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		b.Mu.Unlock()
		c.Next()
	}
}
