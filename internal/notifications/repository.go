package notifications

import (
	"backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

// Repository defines the interface for notification data access
type Repository interface {
	Create(notification *domain.Notification) error
	GetByUserID(userID uint, filters map[string]interface{}) ([]domain.Notification, error)
	GetByID(id uint) (*domain.Notification, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
	GetUnreadCount(userID uint) (int64, error)
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new notification repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *repository) GetByUserID(userID uint, filters map[string]interface{}) ([]domain.Notification, error) {
	var notifications []domain.Notification
	query := r.db.Where("user_id = ?", userID)

	if isRead, ok := filters["is_read"]; ok {
		query = query.Where("is_read = ?", isRead)
	}

	// Apply pagination
	if page, ok := filters["page"].(int); ok {
		limit := 20
		if l, ok := filters["limit"].(int); ok {
			limit = l
		}
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *repository) GetByID(id uint) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *repository) MarkAsRead(id uint, userID uint) error {
	now := time.Now()
	return r.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *repository) MarkAllAsRead(userID uint) error {
	now := time.Now()
	return r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *repository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Notification{}, id).Error
}
