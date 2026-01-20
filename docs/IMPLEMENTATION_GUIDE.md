# Implementation Guide - University Project Hub Backend

## Overview

This guide provides step-by-step instructions for implementing the University Project Hub backend based on the architectural design. Follow the phases sequentially to build a robust, production-ready system.

---

## Prerequisites

### Required Tools

- Go 1.23+
- PostgreSQL 14+
- Git
- Docker (optional, for containerization)
- Postman or similar API testing tool

### Go Dependencies (Already in go.mod)

- `github.com/gin-gonic/gin` - HTTP web framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/spf13/viper` - Configuration management
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/go-playground/validator/v10` - Request validation

---

## Phase 1: Foundation Setup (Week 1)

### 1.1 Update Domain Models

Update [internal/domain/models.go](internal/domain/models.go) to match the enhanced schema:

**Key Changes Needed:**

```go
// Add to Proposal
type Proposal struct {
    // ... existing fields
    SubmissionCount  int                  `json:"submission_count"`
    RejectedAt       *time.Time           `json:"rejected_at"`
    RejectedBy       *uint                `json:"rejected_by"`
    RejectionReason  string               `json:"rejection_reason"`
}

// Add to ProposalVersion
type ProposalVersion struct {
    // ... existing fields
    Methodology      string               `json:"methodology"`
    ExpectedOutcomes string               `json:"expected_outcomes"`
    FileHash         string               `json:"file_hash"`
    FileSizeBytes    int64                `json:"file_size_bytes"`
    IPAddress        string               `json:"ip_address"`
    UserAgent        string               `json:"user_agent"`
    SessionID        string               `json:"session_id"`
}

// Add AuditLog model
type AuditLog struct {
    ID          uint      `gorm:"primaryKey"`
    EntityType  string    `gorm:"type:varchar(50);not null"`
    EntityID    uint
    Action      string    `gorm:"type:varchar(50);not null"`
    ActorID     uint
    ActorRole   string
    ActorEmail  string
    OldState    string    `gorm:"type:jsonb"`
    NewState    string    `gorm:"type:jsonb"`
    Changes     string    `gorm:"type:jsonb"`
    IPAddress   string
    UserAgent   string
    RequestID   string
    SessionID   string
    Timestamp   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
    Metadata    string    `gorm:"type:jsonb"`
}
```

### 1.2 Implement State Machine

Create [internal/proposals/state_machine.go](internal/proposals/state_machine.go):

```go
package proposals

import (
    "backend/internal/domain"
    "backend/pkg/enums"
    "errors"
)

// StateTransition defines allowed state changes
type StateTransition struct {
    From []enums.ProposalStatus
    To   enums.ProposalStatus
}

// ValidTransitions maps all allowed state transitions
var ValidTransitions = map[enums.ProposalStatus][]enums.ProposalStatus{
    enums.ProposalStatusDraft: {
        enums.ProposalStatusSubmitted,
    },
    enums.ProposalStatusSubmitted: {
        enums.ProposalStatusUnderReview,
    },
    enums.ProposalStatusUnderReview: {
        enums.ProposalStatusRevisionRequired,
        enums.ProposalStatusApproved,
        enums.ProposalStatusRejected,
    },
    enums.ProposalStatusRevisionRequired: {
        enums.ProposalStatusDraft, // When new version created
    },
    // Terminal states have no transitions
    enums.ProposalStatusApproved:  {},
    enums.ProposalStatusRejected:  {},
}

// CanTransition checks if state transition is valid
func CanTransition(from, to enums.ProposalStatus) bool {
    allowedStates, exists := ValidTransitions[from]
    if !exists {
        return false
    }

    for _, state := range allowedStates {
        if state == to {
            return true
        }
    }
    return false
}

// ValidateTransition validates and returns error if invalid
func ValidateTransition(from, to enums.ProposalStatus) error {
    if !CanTransition(from, to) {
        return errors.New("invalid state transition from " + string(from) + " to " + string(to))
    }
    return nil
}

// IsEditable checks if proposal can be edited
func IsEditable(status enums.ProposalStatus) bool {
    return status == enums.ProposalStatusDraft
}

// CanCreateVersion checks if new version can be created
func CanCreateVersion(status enums.ProposalStatus) bool {
    return status == enums.ProposalStatusDraft ||
           status == enums.ProposalStatusRevisionRequired
}

// IsTerminal checks if state is terminal (no further transitions)
func IsTerminal(status enums.ProposalStatus) bool {
    return status == enums.ProposalStatusApproved ||
           status == enums.ProposalStatusRejected
}
```

### 1.3 Implement JWT Authentication

Update [internal/auth/jwt.go](internal/auth/jwt.go):

```go
package auth

import (
    "backend/config"
    "backend/internal/domain"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
    UserID       uint   `json:"user_id"`
    Email        string `json:"email"`
    Role         string `json:"role"`
    DepartmentID uint   `json:"department_id"`
    jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token
func GenerateToken(user *domain.User, cfg config.Config) (string, time.Time, error) {
    expirationTime := time.Now().Add(24 * time.Hour)

    claims := &TokenClaims{
        UserID:       user.ID,
        Email:        user.Email,
        Role:         string(user.Role),
        DepartmentID: user.DepartmentID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "university-project-hub",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(cfg.JWTSecret))

    return tokenString, expirationTime, err
}

// ValidateToken validates and parses JWT token
func ValidateToken(tokenString string, cfg config.Config) (*TokenClaims, error) {
    claims := &TokenClaims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(cfg.JWTSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
```

### 1.4 Update Middleware

Update [internal/app/middlewares.go](internal/app/middlewares.go):

```go
package app

import (
    "backend/config"
    "backend/internal/auth"
    "backend/pkg/enums"
    "backend/pkg/response"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Error(c, http.StatusUnauthorized, "Authorization header required", nil)
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if !(len(parts) == 2 && parts[0] == "Bearer") {
            response.Error(c, http.StatusUnauthorized, "Invalid authorization format", nil)
            c.Abort()
            return
        }

        claims, err := auth.ValidateToken(parts[1], cfg)
        if err != nil {
            response.Error(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
            c.Abort()
            return
        }

        // Store user info in context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)
        c.Set("department_id", claims.DepartmentID)

        c.Next()
    }
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...enums.Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            response.Error(c, http.StatusUnauthorized, "User role not found", nil)
            c.Abort()
            return
        }

        role := enums.Role(userRole.(string))
        for _, allowedRole := range allowedRoles {
            if role == allowedRole {
                c.Next()
                return
            }
        }

        response.Error(c, http.StatusForbidden, "Insufficient permissions", nil)
        c.Abort()
    }
}

// AuditMiddleware logs all requests for audit trail
func AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Store request metadata in context for audit logging
        c.Set("ip_address", c.ClientIP())
        c.Set("user_agent", c.GetHeader("User-Agent"))

        // Generate request ID
        requestID := generateRequestID()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}

func generateRequestID() string {
    // Implement UUID generation
    return "req_" + time.Now().Format("20060102150405")
}
```

### 1.5 Create Audit Service

Create [pkg/audit/logger.go](pkg/audit/logger.go):

```go
package audit

import (
    "backend/internal/domain"
    "encoding/json"
    "gorm.io/gorm"
)

type AuditLogger struct {
    db *gorm.DB
}

func NewAuditLogger(db *gorm.DB) *AuditLogger {
    return &AuditLogger{db: db}
}

// Log creates an audit log entry
func (a *AuditLogger) Log(
    entityType string,
    entityID uint,
    action string,
    actorID uint,
    actorRole string,
    oldState interface{},
    newState interface{},
    metadata map[string]interface{},
) error {
    oldJSON, _ := json.Marshal(oldState)
    newJSON, _ := json.Marshal(newState)
    metadataJSON, _ := json.Marshal(metadata)

    log := domain.AuditLog{
        EntityType: entityType,
        EntityID:   entityID,
        Action:     action,
        ActorID:    actorID,
        ActorRole:  actorRole,
        OldState:   string(oldJSON),
        NewState:   string(newJSON),
        Metadata:   string(metadataJSON),
    }

    return a.db.Create(&log).Error
}

// LogProposalSubmission logs proposal submission
func (a *AuditLogger) LogProposalSubmission(
    proposal *domain.Proposal,
    actorID uint,
    actorRole string,
    ipAddress string,
) error {
    metadata := map[string]interface{}{
        "ip_address": ipAddress,
        "version_id": proposal.CurrentVersionID,
    }

    return a.Log(
        "proposal",
        proposal.ID,
        "submit",
        actorID,
        actorRole,
        map[string]interface{}{"status": "draft"},
        map[string]interface{}{"status": "submitted"},
        metadata,
    )
}
```

---

## Phase 2: Authentication Implementation (Days 1-3)

### 2.1 Implement Auth Service

Create [internal/auth/service.go](internal/auth/service.go):

```go
package auth

import (
    "backend/config"
    "backend/internal/domain"
    "errors"
    "golang.org/x/crypto/bcrypt"
)

type Service struct {
    repo   Repository
    config config.Config
}

func NewService(r Repository, cfg config.Config) *Service {
    return &Service{repo: r, config: cfg}
}

// Register creates a new user
func (s *Service) Register(user *domain.User) error {
    // Check if email already exists
    existing, _ := s.repo.FindByEmail(user.Email)
    if existing != nil {
        return errors.New("email already registered")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(user.Password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return err
    }
    user.Password = string(hashedPassword)

    // Create user
    return s.repo.Create(user)
}

// Login authenticates user and returns token
func (s *Service) Login(email, password string) (string, *domain.User, time.Time, error) {
    user, err := s.repo.FindByEmail(email)
    if err != nil {
        return "", nil, time.Time{}, errors.New("invalid credentials")
    }

    // Check if account is active
    if !user.IsActive {
        return "", nil, time.Time{}, errors.New("account is deactivated")
    }

    // Check password
    err = bcrypt.CompareHashAndPassword(
        []byte(user.Password),
        []byte(password),
    )
    if err != nil {
        // Increment failed login attempts
        s.repo.IncrementFailedLogins(user.ID)
        return "", nil, time.Time{}, errors.New("invalid credentials")
    }

    // Reset failed login attempts
    s.repo.ResetFailedLogins(user.ID)

    // Generate token
    token, expiresAt, err := GenerateToken(user, s.config)
    if err != nil {
        return "", nil, time.Time{}, err
    }

    // Update last login
    s.repo.UpdateLastLogin(user.ID)

    return token, user, expiresAt, nil
}
```

### 2.2 Implement Auth Repository

Create [internal/auth/repository.go](internal/auth/repository.go):

```go
package auth

import (
    "backend/internal/domain"
    "time"
    "gorm.io/gorm"
)

type Repository interface {
    Create(user *domain.User) error
    FindByEmail(email string) (*domain.User, error)
    FindByID(id uint) (*domain.User, error)
    IncrementFailedLogins(userID uint) error
    ResetFailedLogins(userID uint) error
    UpdateLastLogin(userID uint) error
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) Create(user *domain.User) error {
    return r.db.Create(user).Error
}

func (r *repository) FindByEmail(email string) (*domain.User, error) {
    var user domain.User
    err := r.db.Where("email = ? AND deleted_at IS NULL", email).
        Preload("University").
        Preload("Department").
        First(&user).Error
    return &user, err
}

func (r *repository) FindByID(id uint) (*domain.User, error) {
    var user domain.User
    err := r.db.Where("id = ? AND deleted_at IS NULL", id).
        Preload("University").
        Preload("Department").
        First(&user).Error
    return &user, err
}

func (r *repository) IncrementFailedLogins(userID uint) error {
    return r.db.Model(&domain.User{}).
        Where("id = ?", userID).
        UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + 1")).
        Error
}

func (r *repository) ResetFailedLogins(userID uint) error {
    return r.db.Model(&domain.User{}).
        Where("id = ?", userID).
        Update("failed_login_attempts", 0).
        Error
}

func (r *repository) UpdateLastLogin(userID uint) error {
    return r.db.Model(&domain.User{}).
        Where("id = ?", userID).
        Updates(map[string]interface{}{
            "last_login_at": time.Now(),
        }).
        Error
}
```

### 2.3 Implement Auth Handlers

Update [internal/auth/handler.go](internal/auth/handler.go):

```go
package auth

import (
    "backend/internal/domain"
    "backend/pkg/response"
    "net/http"

    "github.com/gin-gonic/gin"
)

type Handler struct {
    service *Service
}

func NewHandler(s *Service) *Handler {
    return &Handler{service: s}
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
    rg.POST("/register", h.Register)
    rg.POST("/login", h.Login)
}

type RegisterRequest struct {
    Name         string `json:"name" binding:"required,min=2,max=100"`
    Email        string `json:"email" binding:"required,email"`
    Password     string `json:"password" binding:"required,min=8"`
    Role         string `json:"role" binding:"required,oneof=student teacher admin"`
    UniversityID uint   `json:"university_id" binding:"required"`
    DepartmentID uint   `json:"department_id" binding:"required"`
    StudentID    string `json:"student_id"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
        return
    }

    // Validate student_id for students
    if req.Role == "student" && req.StudentID == "" {
        response.Error(c, http.StatusBadRequest, "Student ID required for students", nil)
        return
    }

    user := &domain.User{
        Name:         req.Name,
        Email:        req.Email,
        Password:     req.Password,
        Role:         enums.Role(req.Role),
        UniversityID: req.UniversityID,
        DepartmentID: req.DepartmentID,
        StudentID:    req.StudentID,
    }

    err := h.service.Register(user)
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error(), nil)
        return
    }

    // Generate token
    token, expiresAt, err := GenerateToken(user, h.service.config)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to generate token", err.Error())
        return
    }

    response.JSON(c, http.StatusCreated, "Registration successful", gin.H{
        "user":       user,
        "token":      token,
        "expires_at": expiresAt,
    })
}

func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
        return
    }

    token, user, expiresAt, err := h.service.Login(req.Email, req.Password)
    if err != nil {
        response.Error(c, http.StatusUnauthorized, err.Error(), nil)
        return
    }

    response.JSON(c, http.StatusOK, "Login successful", gin.H{
        "user":       user,
        "token":      token,
        "expires_at": expiresAt,
    })
}
```

---

## Phase 3: Team Management (Days 4-6)

### 3.1 Implement Team Service

Create comprehensive team service with all business logic:

```go
package teams

import (
    "backend/internal/domain"
    "backend/pkg/enums"
    "errors"
)

type Service struct {
    repo Repository
}

func NewService(r Repository) *Service {
    return &Service{repo: r}
}

// CreateTeam creates a new team with validations
func (s *Service) CreateTeam(team *domain.Team, memberIDs []uint) error {
    // Validate team size (max 5 members including leader)
    if len(memberIDs) > 4 {
        return errors.New("maximum 5 members allowed (including leader)")
    }

    // Check if advisor is a teacher
    advisor, err := s.repo.GetUser(team.AdvisorID)
    if err != nil || advisor.Role != enums.RoleTeacher {
        return errors.New("advisor must be a teacher")
    }

    // Check if advisor is in same department
    if advisor.DepartmentID != team.DepartmentID {
        return errors.New("advisor must be from the same department")
    }

    // Create team with pending status
    team.Status = enums.TeamStatusPendingAdvisorApproval
    err = s.repo.CreateTeam(team, memberIDs)
    if err != nil {
        return err
    }

    // Send notifications
    // - To members: invitation
    // - To advisor: approval request

    return nil
}
```

Continue implementing teams, proposals, and projects following the architecture...

---

## Phase 4-10: Implementation Checklist

### ✅ Completed

- [x] Architecture documentation
- [x] API specification
- [x] Database schema
- [x] Implementation guide structure

### 🚧 To Implement

#### Week 1: Foundation

- [ ] Update domain models with all fields
- [ ] Implement state machine
- [ ] Complete JWT authentication
- [ ] Audit logging system
- [ ] Updated middleware

#### Week 2: Core Features

- [ ] Team management (create, invite, approve)
- [ ] Proposal CRUD operations
- [ ] Version management (immutable)
- [ ] State transitions with validation

#### Week 3: Workflow

- [ ] Proposal submission logic
- [ ] Teacher review and feedback
- [ ] Revision workflow
- [ ] Approval and project creation

#### Week 4: Projects

- [ ] Project creation on approval
- [ ] Documentation upload
- [ ] Publication workflow
- [ ] Public archive

#### Week 5: AI Integration

- [ ] Proposal analysis endpoint
- [ ] Similarity checking
- [ ] Advisory (non-authoritative)

#### Week 6: Polish

- [ ] Notification system
- [ ] Admin endpoints
- [ ] Comprehensive testing
- [ ] Security hardening

---

## Testing Strategy

### Unit Tests

```go
// Example: Test state machine
func TestStateTransitions(t *testing.T) {
    // Test valid transition
    assert.True(t, CanTransition(
        enums.ProposalStatusDraft,
        enums.ProposalStatusSubmitted,
    ))

    // Test invalid transition
    assert.False(t, CanTransition(
        enums.ProposalStatusDraft,
        enums.ProposalStatusApproved,
    ))
}
```

### Integration Tests

- Test full workflows (create team → submit proposal → approve)
- Test authorization at each step
- Test state transitions with database

### API Tests

- Use Postman collections
- Test all endpoints with various roles
- Test edge cases and error conditions

---

## Deployment

### Environment Variables (.env)

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=university_hub
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-change-in-production
ENVIRONMENT=development
```

### Docker Compose

```yaml
version: "3.8"
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_DB: university_hub
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: yourpassword
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
    environment:
      DB_HOST: db

volumes:
  postgres_data:
```

---

## Next Steps

1. **Start with Phase 1**: Set up foundation (models, auth, middleware)
2. **Implement iteratively**: One module at a time, test thoroughly
3. **Follow the architecture**: Reference ARCHITECTURE.md for design decisions
4. **Test continuously**: Write tests as you implement
5. **Document as you go**: Update API docs with actual implementation

This is a solid, well-architected system. Take it step by step, and you'll build something impressive!
