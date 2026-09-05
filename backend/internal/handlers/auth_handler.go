package handlers

import (
	"time"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/dto"
	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Auth *service.AuthService
	cfg  config.Config
}

func NewAuthHandler(auth *service.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{Auth: auth, cfg: cfg}
}

// Register godoc
// @Summary Register a new end user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var in dto.RegisterRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.Email == "" || len(in.Password) < 8 {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	hash, err := h.Auth.HashPassword(in.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}
	user := &models.User{Name: in.Name, Email: in.Email, PasswordHash: hash, Role: "end_user", IsActive: true}
	if err := h.Auth.Users.Create(c.Request.Context(), user); err != nil {
		c.JSON(409, gin.H{"error": "email already registered"})
		return
	}
	c.JSON(201, userResponse(user))
}

// Login godoc
// @Summary Login and issue access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var in dto.LoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	user, err := h.Auth.Users.FindByEmail(c.Request.Context(), in.Email)
	if err != nil || !user.IsActive || !h.Auth.CheckPassword(user.PasswordHash, in.Password) {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	access, err := h.Auth.AccessToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to sign token"})
		return
	}
	refresh, hash, expires, err := h.Auth.NewRefreshToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create refresh token"})
		return
	}
	if err := h.Auth.RefreshRepo.Create(c.Request.Context(), user.ID, hash, expires); err != nil {
		c.JSON(500, gin.H{"error": "failed to persist refresh token"})
		return
	}
	h.setRefreshCookie(c, refresh, expires)
	resp := dto.LoginResponse{AccessToken: access, User: userResponse(user)}
	if h.cfg.AuthDevTokenBody {
		resp.RefreshToken = refresh
	}
	c.JSON(200, resp)
}

// Refresh godoc
// @Summary Refresh the access token
// @Tags auth
// @Produce json
// @Security RefreshCookie
// @Success 200 {object} dto.RefreshResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	plain, _ := c.Cookie("refresh_token")
	if plain == "" && h.cfg.AuthDevTokenBody {
		var in dto.RefreshRequest
		_ = c.ShouldBindJSON(&in)
		plain = in.RefreshToken
	}
	if plain == "" {
		c.JSON(401, gin.H{"error": "missing refresh token"})
		return
	}
	hash := h.Auth.HashRefreshToken(plain)
	row, err := h.Auth.RefreshRepo.FindValidByHash(c.Request.Context(), hash, time.Now())
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}
	user, err := h.Auth.Users.FindByID(c.Request.Context(), row.UserID)
	if err != nil || !user.IsActive {
		c.JSON(401, gin.H{"error": "invalid user"})
		return
	}
	access, err := h.Auth.AccessToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to sign token"})
		return
	}
	c.JSON(200, dto.RefreshResponse{AccessToken: access})
}

// Logout godoc
// @Summary Revoke the refresh token and clear its cookie
// @Tags auth
// @Produce json
// @Security RefreshCookie
// @Success 200 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	plain, _ := c.Cookie("refresh_token")
	if plain == "" && h.cfg.AuthDevTokenBody {
		var in dto.RefreshRequest
		_ = c.ShouldBindJSON(&in)
		plain = in.RefreshToken
	}
	if plain != "" {
		_ = h.Auth.RefreshRepo.RevokeByHash(c.Request.Context(), h.Auth.HashRefreshToken(plain))
	}
	h.clearRefreshCookie(c)
	c.JSON(200, gin.H{"message": "logged out"})
}

// Me godoc
// @Summary Get the authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 404 {object} map[string]string
// @Router /api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.Auth.Users.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, userResponse(user))
}

func userResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, expires time.Time) {
	maxAge := int(time.Until(expires).Seconds())
	c.SetCookie("refresh_token", token, maxAge, "/api/auth", "", h.cfg.AppEnv == "production", true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/api/auth", "", h.cfg.AppEnv == "production", true)
}
