package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/repository"
	"helpdesk-backend/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	name := strings.TrimSpace(os.Getenv("SUPER_ADMIN_NAME"))
	email := strings.TrimSpace(os.Getenv("SUPER_ADMIN_EMAIL"))
	password := os.Getenv("SUPER_ADMIN_PASSWORD")

	if name == "" || email == "" || password == "" {
		log.Fatal("SUPER_ADMIN_NAME, SUPER_ADMIN_EMAIL, and SUPER_ADMIN_PASSWORD are required")
	}
	if len(password) < 8 {
		log.Fatal("SUPER_ADMIN_PASSWORD must be at least 8 characters")
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authSvc := &service.AuthService{}
	hash, err := authSvc.HashPassword(password)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}

	users := repository.NewUserRepository(db)
	if err := users.UpsertBootstrapSuperAdmin(ctx, user); err != nil {
		log.Fatalf("failed to upsert super admin: %v", err)
	}

	log.Printf("super admin ready: id=%d email=%s", user.ID, user.Email)
}
