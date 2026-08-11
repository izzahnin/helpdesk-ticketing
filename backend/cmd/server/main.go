package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"helpdesk-backend/internal/config"
)

func main() {
	cfg := config.Load()

	router := gin.Default()
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
