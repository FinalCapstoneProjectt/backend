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

// ForgotPassword initiates password reset
// @Summary Request password reset
// @Description Sends a password reset link to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} response.Response
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resetToken, err := h.service.ForgotPassword(req.Email)
	if err != nil {
		// Don't reveal error to prevent email enumeration
		response.JSON(c, http.StatusOK, "If the email exists, a reset link will be sent", nil)
		return
	}

	// In production, send email instead of returning token
	// For demo/development, return the token
	response.JSON(c, http.StatusOK, "Password reset initiated", gin.H{
		"reset_token": resetToken,
		"note":        "In production, this token would be sent via email",
	})
}

// ResetPassword resets password with token
// @Summary Reset password
// @Description Reset password using the reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.JSON(c, http.StatusOK, "Password reset successfully", nil)
}

// UpdateProfile updates user profile
// @Summary Update profile
// @Description Update the authenticated user's profile
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile data"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.service.UpdateProfile(userID.(uint), req.Name, req.ProfilePhoto)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	response.JSON(c, http.StatusOK, "Profile updated successfully", user)
}

// ChangePassword changes user password
// @Summary Change password
// @Description Change password for authenticated user (requires current password)
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.service.ChangePassword(userID.(uint), req.CurrentPassword, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.JSON(c, http.StatusOK, "Password changed successfully", nil)
}

// Request structs for new endpoints
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UpdateProfileRequest struct {
	Name         string `json:"name"`
	ProfilePhoto string `json:"profile_photo"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
