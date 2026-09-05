package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/routes"
	"helpdesk-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type testApp struct {
	router *gin.Engine
	db     *sqlx.DB
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

func TestAuthAndRoleIntegration(t *testing.T) {
	app := newTestApp(t)

	registerResp := app.request(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "End User",
		"email":    "end.user@test.local",
		"password": "password123",
	})
	assertStatus(t, registerResp, http.StatusCreated)

	var registered struct {
		ID   int64  `json:"id"`
		Role string `json:"role"`
	}
	decodeJSON(t, registerResp, &registered)
	if registered.Role != "end_user" {
		t.Fatalf("registered role = %q, want end_user", registered.Role)
	}

	adminToken := app.login(t, "admin@test.local", "password123").AccessToken
	promoteToSuperAdmin := app.request(t, http.MethodPatch, fmt.Sprintf("/api/users/%d", registered.ID), adminToken, map[string]any{
		"role":      "super_admin",
		"is_active": true,
	})
	assertStatus(t, promoteToSuperAdmin, http.StatusForbidden)

	superToken := app.login(t, "super@test.local", "password123").AccessToken
	promoteToAdmin := app.request(t, http.MethodPatch, fmt.Sprintf("/api/users/%d", registered.ID), superToken, map[string]any{
		"role":      "admin",
		"is_active": true,
	})
	assertStatus(t, promoteToAdmin, http.StatusOK)
}

func TestTicketWorkflowIntegration(t *testing.T) {
	app := newTestApp(t)

	superToken := app.login(t, "super@test.local", "password123").AccessToken
	endUserToken := app.login(t, "requester@test.local", "password123").AccessToken
	staffToken := app.login(t, "staff@test.local", "password123").AccessToken
	categoryID := app.createCategory(t, superToken)
	staffID := app.userIDByEmail(t, "staff@test.local")

	submitResp := app.request(t, http.MethodPost, "/api/tickets", endUserToken, map[string]any{
		"title":       "Laptop tidak bisa connect WiFi",
		"description": "Sejak pagi laptop tidak bisa terhubung ke jaringan kantor.",
		"category_id": categoryID,
		"priority":    "High",
	})
	assertStatus(t, submitResp, http.StatusCreated)

	var ticket struct {
		ID          int64  `json:"id"`
		RequesterID int64  `json:"requester_id"`
		Priority    string `json:"priority"`
		Status      string `json:"status"`
	}
	decodeJSON(t, submitResp, &ticket)
	if ticket.Status != "Open" {
		t.Fatalf("new ticket status = %q, want Open", ticket.Status)
	}

	assignResp := app.request(t, http.MethodPatch, fmt.Sprintf("/api/tickets/%d/assign", ticket.ID), superToken, map[string]any{
		"assignee_id": staffID,
	})
	assertStatus(t, assignResp, http.StatusOK)

	statusResp := app.request(t, http.MethodPatch, fmt.Sprintf("/api/tickets/%d/status", ticket.ID), staffToken, map[string]any{
		"status": "In Progress",
	})
	assertStatus(t, statusResp, http.StatusOK)

	commentResp := app.request(t, http.MethodPost, fmt.Sprintf("/api/tickets/%d/comments", ticket.ID), staffToken, map[string]any{
		"message": "Ticket sedang dicek oleh tim IT.",
	})
	assertStatus(t, commentResp, http.StatusCreated)

	otherUserToken := app.login(t, "other.user@test.local", "password123").AccessToken
	forbiddenResp := app.request(t, http.MethodGet, fmt.Sprintf("/api/tickets/%d", ticket.ID), otherUserToken, nil)
	assertStatus(t, forbiddenResp, http.StatusForbidden)
}

func TestDashboardAndReportIntegration(t *testing.T) {
	app := newTestApp(t)

	superToken := app.login(t, "super@test.local", "password123").AccessToken
	staffToken := app.login(t, "staff@test.local", "password123").AccessToken

	adminSummaryResp := app.request(t, http.MethodGet, "/api/dashboard/summary", superToken, nil)
	assertStatus(t, adminSummaryResp, http.StatusOK)

	staffAdminSummaryResp := app.request(t, http.MethodGet, "/api/dashboard/summary", staffToken, nil)
	assertStatus(t, staffAdminSummaryResp, http.StatusForbidden)

	myPerformanceResp := app.request(t, http.MethodGet, "/api/dashboard/my-performance", staffToken, nil)
	assertStatus(t, myPerformanceResp, http.StatusOK)

	staffIDFilterResp := app.request(t, http.MethodGet, "/api/dashboard/my-performance?staff_id=1", staffToken, nil)
	assertStatus(t, staffIDFilterResp, http.StatusBadRequest)

	reportResp := app.request(t, http.MethodGet, "/api/reports/export", superToken, nil)
	assertStatus(t, reportResp, http.StatusOK)
	if contentType := reportResp.Header().Get("Content-Type"); !strings.Contains(contentType, "text/csv") {
		t.Fatalf("Content-Type = %q, want text/csv", contentType)
	}
	if body := reportResp.Body.String(); !strings.Contains(body, "ticket_id,title,category,priority,status,requester,assignee,created_at,resolved_at,sla_deadline,sla_state") {
		t.Fatalf("CSV body does not contain expected header: %q", body)
	}
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()

	backendRoot := findBackendRoot(t)
	_ = godotenv.Load(filepath.Join(backendRoot, ".env"))

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	guardTestDatabase(t, dsn)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	resetDatabase(t, db, backendRoot)
	seedUsers(t, db)
	middleware.ResetRateLimitBuckets()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.Register(router, db, config.Config{
		AppEnv:             "test",
		Port:               "0",
		DatabaseURL:        dsn,
		JWTAccessSecret:    "integration-test-secret",
		RefreshTokenPepper: "integration-test-pepper",
		FrontendURL:        "http://localhost:3000",
		AuthDevTokenBody:   true,
	})

	return &testApp{router: router, db: db}
}

func guardTestDatabase(t *testing.T, dsn string) {
	t.Helper()

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if !strings.Contains(strings.ToLower(dbName), "test") {
		t.Fatalf("TEST_DATABASE_URL database name must contain 'test', got %q", dbName)
	}
}

func findBackendRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find backend root containing go.mod")
		}
		dir = parent
	}
}

func resetDatabase(t *testing.T, db *sqlx.DB, backendRoot string) {
	t.Helper()

	_, err := db.Exec(`
		DROP TABLE IF EXISTS
			ticket_activity_log,
			ticket_status_log,
			ticket_comments,
			tickets,
			refresh_tokens,
			sla_rules,
			categories,
			users
		CASCADE
	`)
	if err != nil {
		t.Fatalf("drop test tables: %v", err)
	}

	for _, migration := range []string{"001_init.sql", "002_add_super_admin_role.sql"} {
		path := filepath.Join(backendRoot, "migration", migration)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", migration, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			t.Fatalf("run migration %s: %v", migration, err)
		}
	}
}

func seedUsers(t *testing.T, db *sqlx.DB) {
	t.Helper()

	authSvc := &service.AuthService{}
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	users := []struct {
		name  string
		email string
		role  string
	}{
		{name: "Super Admin", email: "super@test.local", role: "super_admin"},
		{name: "Admin", email: "admin@test.local", role: "admin"},
		{name: "Staff", email: "staff@test.local", role: "staff"},
		{name: "Requester", email: "requester@test.local", role: "end_user"},
		{name: "Other User", email: "other.user@test.local", role: "end_user"},
	}

	for _, user := range users {
		_, err := db.Exec(`
			INSERT INTO users (name, email, password_hash, role, is_active)
			VALUES ($1,$2,$3,$4,true)
		`, user.name, user.email, hash, user.role)
		if err != nil {
			t.Fatalf("seed user %s: %v", user.email, err)
		}
	}
}

func (app *testApp) createCategory(t *testing.T, token string) int64 {
	t.Helper()

	resp := app.request(t, http.MethodPost, "/api/categories", token, map[string]any{
		"name":              "Network",
		"default_sla_hours": 24,
	})
	assertStatus(t, resp, http.StatusCreated)

	var category struct {
		ID int64 `json:"id"`
	}
	decodeJSON(t, resp, &category)
	return category.ID
}

func (app *testApp) userIDByEmail(t *testing.T, email string) int64 {
	t.Helper()

	var id int64
	if err := app.db.Get(&id, `SELECT id FROM users WHERE email=$1`, email); err != nil {
		t.Fatalf("find user %s: %v", email, err)
	}
	return id
}

func (app *testApp) login(t *testing.T, email, password string) loginResponse {
	t.Helper()

	resp := app.request(t, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	assertStatus(t, resp, http.StatusOK)

	var out loginResponse
	decodeJSON(t, resp, &out)
	if out.AccessToken == "" {
		t.Fatal("login response has empty access_token")
	}
	return out
}

func (app *testApp) request(t *testing.T, method, path, accessToken string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp := httptest.NewRecorder()
	app.router.ServeHTTP(resp, req)
	return resp
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", resp.Code, want, resp.Body.String())
	}
}

func decodeJSON(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v; body = %s", err, resp.Body.String())
	}
}
