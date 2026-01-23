package notifications

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(notification *domain.Notification) error
	GetByID(id uint) (*domain.Notification, error)
	GetByUserID(userID uint) ([]domain.Notification, error)
	GetUnreadByUserID(userID uint) ([]domain.Notification, error)
	GetUnreadCount(userID uint) (int64, error)
	MarkAsRead(id uint) error
	MarkAllAsRead(userID uint) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *repository) GetByID(id uint) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *repository) GetByUserID(userID uint) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *repository) GetUnreadByUserID(userID uint) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("user_id = ? AND is_read = ?", userID, false).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *repository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *repository) MarkAsRead(id uint) error {
	return r.db.Model(&domain.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *repository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&domain.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Notification{}, id).Error
}
