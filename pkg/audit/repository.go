package audit

import (
	"backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

// Repository defines the interface for audit log data access
type Repository interface {
	GetLogs(filters AuditFilters) ([]domain.AuditLog, int64, error)
	GetByID(id uint) (*domain.AuditLog, error)
}

// AuditFilters contains filter options for querying audit logs
type AuditFilters struct {
	EntityType string
	EntityID   uint
	ActorID    uint
	Action     string
	FromDate   *time.Time
	ToDate     *time.Time
	Page       int
	Limit      int
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetLogs(filters AuditFilters) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.Model(&domain.AuditLog{})

	// Apply filters
	if filters.EntityType != "" {
		query = query.Where("entity_type = ?", filters.EntityType)
	}
	if filters.EntityID > 0 {
		query = query.Where("entity_id = ?", filters.EntityID)
	}
	if filters.ActorID > 0 {
		query = query.Where("actor_id = ?", filters.ActorID)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.FromDate != nil {
		query = query.Where("timestamp >= ?", filters.FromDate)
	}
	if filters.ToDate != nil {
		query = query.Where("timestamp <= ?", filters.ToDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Fetch logs with actor preload
	err := query.
		Preload("Actor").
		Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	return logs, total, err
}

func (r *repository) GetByID(id uint) (*domain.AuditLog, error) {
	var log domain.AuditLog
	err := r.db.Preload("Actor").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
