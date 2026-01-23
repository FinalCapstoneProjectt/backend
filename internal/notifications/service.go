package notifications

import (
	"backend/internal/domain"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) GetUserNotifications(userID uint) ([]domain.Notification, error) {
	return s.repo.GetByUserID(userID)
}

func (s *Service) GetUnreadNotifications(userID uint) ([]domain.Notification, error) {
	return s.repo.GetUnreadByUserID(userID)
}

func (s *Service) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

func (s *Service) MarkAsRead(id uint, userID uint) error {
	notification, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if notification.UserID != userID {
		return nil // silently ignore if not the owner
	}
	return s.repo.MarkAsRead(id)
}

func (s *Service) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *Service) DeleteNotification(id uint, userID uint) error {
	notification, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if notification.UserID != userID {
		return nil // silently ignore if not the owner
	}
	return s.repo.Delete(id)
}

func (s *Service) CreateNotification(userID uint, refType string, refID uint, title, message, actionURL string) error {
	notification := &domain.Notification{
		UserID:        userID,
		ReferenceType: refType,
		ReferenceID:   refID,
		Title:         title,
		Message:       message,
		ActionURL:     actionURL,
		Priority:      "normal",
	}
	return s.repo.Create(notification)
}
