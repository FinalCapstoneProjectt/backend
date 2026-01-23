package auth

import (
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new account. Role must be: student, advisor, admin, or public.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			response.Error(c, http.StatusConflict, err.Error(), err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	// Don't expose password
	user.Password = ""

	response.JSON(c, http.StatusCreated, "User registered successfully", user)
}

// Login handles user login
// @Summary Login user
// @Description Authenticates a user and returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get request context
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	loginResp, err := h.service.Login(req, ipAddress, userAgent, requestID)
	if err != nil {
		if err.Error() == "account is temporarily locked due to too many failed login attempts" {
			response.Error(c, http.StatusForbidden, err.Error(), err)
			return
		}
		response.Error(c, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	response.JSON(c, http.StatusOK, "Login successful", loginResp)
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Invalidates old token (optional) and issues a new one.
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		response.Error(c, http.StatusUnauthorized, "Invalid authorization header", nil)
		return
	}

	tokenString = tokenString[7:] // Remove "Bearer " prefix

	newToken, expiresAt, err := h.service.RefreshToken(tokenString)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Failed to refresh token", err)
		return
	}

	response.JSON(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"token":      newToken,
		"expires_at": expiresAt,
	})
}

// GetProfile returns the authenticated user's profile
// @Summary Get user profile
// @Description Returns the profile of the currently logged-in user.
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// User info is set by AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")
	departmentID, _ := c.Get("department_id")

	response.JSON(c, http.StatusOK, "Profile retrieved successfully", gin.H{
		"id":            userID,
		"email":         email,
		"role":          role,
		"department_id": departmentID,
	})
}
