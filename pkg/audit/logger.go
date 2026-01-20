package audit

import (
	"backend/internal/domain"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Logger struct {
	db *gorm.DB
}

func NewLogger(db *gorm.DB) *Logger {
	return &Logger{db: db}
}

// Log creates a generic audit log entry
func (a *Logger) Log(log *domain.AuditLog) error {
	return a.db.Create(log).Error
}

// LogAction creates an audit log with basic information
func (a *Logger) LogAction(
	entityType string,
	entityID uint,
	action string,
	actorID *uint,
	actorRole string,
	actorEmail string,
	oldState interface{},
	newState interface{},
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	oldJSON, _ := json.Marshal(oldState)
	newJSON, _ := json.Marshal(newState)

	log := &domain.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		ActorRole:  actorRole,
		ActorEmail: actorEmail,
		OldState:   string(oldJSON),
		NewState:   string(newJSON),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		SessionID:  sessionID,
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// LogProposalSubmission logs proposal submission with full context
func (a *Logger) LogProposalSubmission(
	proposalID uint,
	versionID uint,
	actorID uint,
	actorRole string,
	actorEmail string,
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	metadata := map[string]interface{}{
		"version_id":       versionID,
		"action_timestamp": time.Now(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	log := &domain.AuditLog{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "submit",
		ActorID:    &actorID,
		ActorRole:  actorRole,
		ActorEmail: actorEmail,
		OldState:   `{"status":"draft"}`,
		NewState:   `{"status":"submitted"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		SessionID:  sessionID,
		Metadata:   string(metadataJSON),
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// LogProposalApproval logs proposal approval
func (a *Logger) LogProposalApproval(
	proposalID uint,
	versionID uint,
	actorID uint,
	actorRole string,
	actorEmail string,
	feedbackID uint,
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	metadata := map[string]interface{}{
		"version_id":  versionID,
		"feedback_id": feedbackID,
		"decision":    "approve",
	}
	metadataJSON, _ := json.Marshal(metadata)

	log := &domain.AuditLog{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "approve",
		ActorID:    &actorID,
		ActorRole:  actorRole,
		ActorEmail: actorEmail,
		OldState:   `{"status":"under_review"}`,
		NewState:   `{"status":"approved"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		SessionID:  sessionID,
		Metadata:   string(metadataJSON),
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// LogTeamCreation logs team creation
func (a *Logger) LogTeamCreation(
	teamID uint,
	leaderID uint,
	actorEmail string,
	memberIDs []uint,
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	metadata := map[string]interface{}{
		"member_ids": memberIDs,
		"leader_id":  leaderID,
	}
	metadataJSON, _ := json.Marshal(metadata)

	log := &domain.AuditLog{
		EntityType: "team",
		EntityID:   teamID,
		Action:     "create",
		ActorID:    &leaderID,
		ActorRole:  "student",
		ActorEmail: actorEmail,
		NewState:   `{"status":"pending_advisor_approval"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		SessionID:  sessionID,
		Metadata:   string(metadataJSON),
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// LogUserLogin logs user login attempt
func (a *Logger) LogUserLogin(
	userID uint,
	email string,
	role string,
	success bool,
	ipAddress string,
	userAgent string,
	requestID string,
) error {
	action := "login_success"
	if !success {
		action = "login_failed"
	}

	log := &domain.AuditLog{
		EntityType: "user",
		EntityID:   userID,
		Action:     action,
		ActorID:    &userID,
		ActorRole:  role,
		ActorEmail: email,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// LogVersionCreation logs proposal version creation
func (a *Logger) LogVersionCreation(
	proposalID uint,
	versionID uint,
	versionNumber int,
	actorID uint,
	actorEmail string,
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	metadata := map[string]interface{}{
		"version_id":     versionID,
		"version_number": versionNumber,
	}
	metadataJSON, _ := json.Marshal(metadata)

	log := &domain.AuditLog{
		EntityType: "proposal_version",
		EntityID:   versionID,
		Action:     "create",
		ActorID:    &actorID,
		ActorRole:  "student",
		ActorEmail: actorEmail,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		SessionID:  sessionID,
		Metadata:   string(metadataJSON),
		Timestamp:  time.Now(),
	}

	return a.db.Create(log).Error
}

// GetAuditLogs retrieves audit logs with filtering
func (a *Logger) GetAuditLogs(filters map[string]interface{}, limit int, offset int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := a.db.Model(&domain.AuditLog{})

	// Apply filters
	if entityType, ok := filters["entity_type"].(string); ok && entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID, ok := filters["entity_id"].(uint); ok && entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	if actorID, ok := filters["actor_id"].(uint); ok && actorID > 0 {
		query = query.Where("actor_id = ?", actorID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if fromDate, ok := filters["from_date"].(time.Time); ok {
		query = query.Where("timestamp >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"].(time.Time); ok {
		query = query.Where("timestamp <= ?", toDate)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Preload("Actor").
		Find(&logs).Error

	return logs, total, err
}
