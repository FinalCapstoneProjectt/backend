package auth

import (
	"backend/config"
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID       uint       `json:"user_id"`
	Email        string     `json:"email"`
	Role         enums.Role `json:"role"`
	DepartmentID uint       `json:"department_id"`
	UniversityID uint       `json:"university_id"` 
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *domain.User, cfg config.Config) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &TokenClaims{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		DepartmentID: user.DepartmentID,
		UniversityID: user.UniversityID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "university-project-hub",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string, cfg config.Config) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// RefreshToken generates a new token if the old one is close to expiring
func RefreshToken(oldTokenString string, cfg config.Config) (string, time.Time, error) {
	claims, err := ValidateToken(oldTokenString, cfg)
	if err != nil {
		return "", time.Time{}, err
	}

	// Only refresh if token is within 1 hour of expiring
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", time.Time{}, errors.New("token does not need refresh yet")
	}

	// Create new token with same claims
	expirationTime := time.Now().Add(24 * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign new token: %w", err)
	}
	return tokenString, expirationTime, err
}

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: secret}
}

func (j *JWTService) GenerateToken(userID int, role string) (string, error) {
	// TODO: Implement token generation
	return "", nil
}
