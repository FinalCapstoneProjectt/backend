package auth

import (
	"backend/config"
	"backend/internal/domain"
	"backend/pkg/audit"
	"backend/pkg/enums"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(req RegisterRequest) (*domain.User, error)
	Login(req LoginRequest, ipAddress string, userAgent string, requestID string) (*LoginResponse, error)
	ValidateToken(token string) (*TokenClaims, error)
	RefreshToken(token string) (string, time.Time, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(token string, newPassword string) error
	UpdateProfile(userID uint, name string, profilePhoto string) (*domain.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
}

type service struct {
	repo        Repository
	cfg         config.Config
	auditLogger *audit.Logger
}

func NewService(repo Repository, cfg config.Config, auditLogger *audit.Logger) Service {
	return &service{
		repo:        repo,
		cfg:         cfg,
		auditLogger: auditLogger,
	}
}

type RegisterRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	Role         string `json:"role" binding:"required" example:"student"` // Swagger example
	UniversityID uint   `json:"university_id" binding:"required"`
	DepartmentID uint   `json:"department_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

// Register creates a new user account
func (s *service) Register(req RegisterRequest) (*domain.User, error) {
	
	// Strict Role validation
	if !enums.IsValidRole(req.Role) {
		return nil, errors.New("invalid role: must be 'student', 'advisor', 'admin', or 'public'")
	}

	// Check if user already exists
	existingUser, err := s.repo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &domain.User{
		Name:                req.Name,
		Email:               req.Email,
		Password:            string(hashedPassword),
		Role:                enums.Role(req.Role),
		UniversityID:        req.UniversityID,
		DepartmentID:        req.DepartmentID,
		EmailVerified:       false,
		FailedLoginAttempts: 0,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *service) Login(req LoginRequest, ipAddress string, userAgent string, requestID string) (*LoginResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		// Log failed login attempt
		s.auditLogger.LogAction("user", 0, "login_failed", nil, "", req.Email, nil, nil, ipAddress, userAgent, requestID, "")
		return nil, errors.New("invalid email or password")
	}

	// Check if account is locked
	locked, err := s.repo.IsAccountLocked(user.ID)
	if err == nil && locked {
		return nil, errors.New("account is temporarily locked due to too many failed login attempts")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Increment failed login attempts
		s.repo.IncrementFailedLogins(user.ID)

		// Lock account if failed attempts exceed threshold (5 attempts)
		if user.FailedLoginAttempts+1 >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			s.repo.LockAccount(user.ID, lockUntil)
		}

		// Log failed login
		s.auditLogger.LogUserLogin(user.ID, user.Email, string(user.Role), false, ipAddress, userAgent, requestID)
		return nil, errors.New("invalid email or password")
	}

	// Reset failed login attempts on successful login
	s.repo.ResetFailedLogins(user.ID)

	// Update last login timestamp
	s.repo.UpdateLastLogin(user.ID)

	// Generate JWT token
	token, expiresAt, err := GenerateToken(user, s.cfg)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Log successful login
	s.auditLogger.LogUserLogin(user.ID, user.Email, string(user.Role), true, ipAddress, userAgent, requestID)

	// Don't expose password in response
	user.Password = ""

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *service) ValidateToken(token string) (*TokenClaims, error) {
	return ValidateToken(token, s.cfg)
}

// RefreshToken generates a new token if the current one is expiring soon
func (s *service) RefreshToken(token string) (string, time.Time, error) {
	return RefreshToken(token, s.cfg)
}

// ForgotPassword generates a password reset token (mock - would normally send email)
func (s *service) ForgotPassword(email string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not
		return "", nil
	}

	// Generate a reset token (in production, store this and send via email)
	resetToken, _, err := GenerateToken(user, s.cfg)
	if err != nil {
		return "", errors.New("failed to generate reset token")
	}

	// Log the action
	s.auditLogger.LogAction("user", user.ID, "password_reset_requested", nil, string(user.Role), user.Email, nil, nil, "", "", "", "")

	return resetToken, nil
}

// ResetPassword resets user password with a valid token
func (s *service) ResetPassword(token string, newPassword string) error {
	// Validate the reset token
	claims, err := ValidateToken(token, s.cfg)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password
	return s.repo.UpdatePassword(claims.UserID, string(hashedPassword))
}

// UpdateProfile updates user profile information
func (s *service) UpdateProfile(userID uint, name string, profilePhoto string) (*domain.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if name != "" {
		user.Name = name
	}
	if profilePhoto != "" {
		user.ProfilePhoto = profilePhoto
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

// ChangePassword changes user password (requires old password)
func (s *service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.repo.UpdatePassword(userID, string(hashedPassword))
}
