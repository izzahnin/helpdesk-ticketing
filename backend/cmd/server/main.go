package main

import (
	"log"

	_ "helpdesk-backend/docs"
	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Helpdesk Ticketing API
// @version 1.0
// @description REST API for the internal helpdesk and ticketing system.
// @host localhost:4000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token using the Bearer scheme.
// @securityDefinitions.apikey RefreshCookie
// @in cookie
// @name refresh_token

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	app := gin.Default()
	if err := app.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to configure trusted proxies: %v", err)
	}
	app.Use(middleware.RequestLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.Register(app, db, cfg)
	log.Fatal(app.Run(":" + cfg.Port))
}
