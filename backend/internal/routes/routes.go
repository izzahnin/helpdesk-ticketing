package routes

import (
	"time"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/handlers"
	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/repository"
	"helpdesk-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Register(app *gin.Engine, db *sqlx.DB, cfg config.Config) {
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	reportRepo := repository.NewReportRepository(db)

	authSvc := &service.AuthService{Users: userRepo, RefreshRepo: refreshRepo, AccessSecret: cfg.JWTAccessSecret, Pepper: cfg.RefreshTokenPepper}
	slaSvc := service.NewSLAService(categoryRepo)
	ticketSvc := service.NewTicketService(ticketRepo, slaSvc)

	authH := handlers.NewAuthHandler(authSvc, cfg)
	userH := &handlers.UserHandler{Users: userRepo}
	categoryH := &handlers.CategoryHandler{Categories: categoryRepo}
	ticketH := &handlers.TicketHandler{Tickets: ticketSvc, Repo: ticketRepo}
	dashboardH := &handlers.DashboardHandler{Repo: dashboardRepo}
	reportH := &handlers.ReportHandler{Repo: reportRepo}
	systemH := handlers.NewSystemHandler(db)

	app.GET("/api/health", systemH.Health)
	app.GET("/api/ready", systemH.Ready)

	app.POST("/api/auth/register", middleware.RateLimit("register", 10, time.Hour), authH.Register)
	app.POST("/api/auth/login", middleware.RateLimit("login", 5, 15*time.Minute), authH.Login)
	app.POST("/api/auth/refresh", middleware.RateLimit("refresh", 30, 15*time.Minute), authH.Refresh)
	app.POST("/api/auth/logout", authH.Logout)

	auth := middleware.RequireAuth(cfg.JWTAccessSecret)
	api := app.Group("/api", auth)
	api.GET("/auth/me", authH.Me)
	api.GET("/categories", categoryH.List)
	api.GET("/tickets", ticketH.List)
	api.GET("/tickets/:id", ticketH.Detail)
	api.POST("/tickets", ticketH.Submit)
	api.POST("/tickets/:id/comments", ticketH.AddComment)

	staff := api.Group("", middleware.RequireRole("staff", "admin", "super_admin"))
	staff.PATCH("/tickets/:id/status", ticketH.UpdateStatus)
	staff.GET("/dashboard/my-performance", dashboardH.MyPerformance)

	admin := api.Group("", middleware.RequireRole("admin", "super_admin"))
	admin.GET("/users", userH.List)
	admin.GET("/users/staff", userH.Staff)
	admin.PATCH("/users/:id", userH.Update)
	admin.POST("/categories", categoryH.Create)
	admin.PATCH("/categories/:id", categoryH.Update)
	admin.DELETE("/categories/:id", categoryH.Delete)
	admin.POST("/sla-rules", categoryH.SetSLARule)
	admin.PATCH("/tickets/:id/assign", ticketH.Assign)
	admin.GET("/dashboard/summary", dashboardH.Summary)
	admin.GET("/dashboard/trend", dashboardH.Trend)
	admin.GET("/dashboard/staff-performance", dashboardH.StaffPerformance)
	admin.GET("/dashboard/sla-breaches", dashboardH.SLABreaches)
	admin.GET("/reports/export", middleware.RateLimit("export", 10, 10*time.Minute), reportH.ExportCSV)
}
