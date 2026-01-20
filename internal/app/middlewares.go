package app

import (
	"backend/config"
	"backend/internal/auth"
	"backend/pkg/audit"
	"backend/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required", nil)
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(tokenString, cfg)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token", err)
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("department_id", claims.DepartmentID)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User role not found", nil)
			c.Abort()
			return
		}

		role := userRole.(string)

		// Check if user role is in allowed roles
		allowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Error(c, http.StatusForbidden, "Insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RBACMiddleware is an alias for RoleMiddleware for backward compatibility
func RBACMiddleware(allowedRoles []string) gin.HandlerFunc {
	return RoleMiddleware(allowedRoles...)
}

// AuditMiddleware logs all requests for audit trail
func AuditMiddleware(auditLogger *audit.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture request start time
		startTime := time.Now()

		// Get request context
		requestID, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")
		userRole, _ := c.Get("user_role")

		// Process request
		c.Next()

		// Log after request completes
		duration := time.Since(startTime)

		// Only log write operations (POST, PUT, DELETE, PATCH)
		if c.Request.Method != "GET" && c.Request.Method != "OPTIONS" {
			var actorID *uint
			if userID != nil {
				id := userID.(uint)
				actorID = &id
			}

			role := ""
			if userRole != nil {
				role = userRole.(string)
			}

			email := ""
			if userEmail != nil {
				email = userEmail.(string)
			}

			reqID := ""
			if requestID != nil {
				reqID = requestID.(string)
			}

			// Log the action
			auditLogger.LogAction(
				"http_request",
				0, // No specific entity ID for general requests
				c.Request.Method+" "+c.Request.URL.Path,
				actorID,
				role,
				email,
				nil, // No old state for HTTP requests
				map[string]interface{}{
					"status_code": c.Writer.Status(),
					"duration_ms": duration.Milliseconds(),
				},
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				reqID,
				"", // Session ID not used for HTTP requests
			)
		}
	}
}

// RateLimitMiddleware implements simple rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	type client struct {
		requests  int
		resetTime time.Time
	}

	clients := make(map[string]*client)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Get or create client entry
		cl, exists := clients[ip]
		if !exists || now.After(cl.resetTime) {
			clients[ip] = &client{
				requests:  1,
				resetTime: now.Add(1 * time.Minute),
			}
			c.Next()
			return
		}

		// Check rate limit (100 requests per minute)
		if cl.requests >= 100 {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		cl.requests++
		c.Next()
	}
}
