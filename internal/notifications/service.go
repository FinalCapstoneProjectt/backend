package notifications

import (
	"backend/internal/domain"
	"errors"
)

// Service handles notification business logic
type Service struct {
	repo Repository
}

// NewService creates a new notification service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateNotification creates a new notification for a user
func (s *Service) CreateNotification(userID uint, refType string, refID uint, title, message, actionURL string) error {
	notification := &domain.Notification{
		UserID:        userID,
		ReferenceType: refType,
		ReferenceID:   refID,
		Title:         title,
		Message:       message,
		ActionURL:     actionURL,
		IsRead:        false,
		Priority:      "normal",
	}

	return s.repo.Create(notification)
}

// CreateNotificationWithPriority creates a notification with specified priority
func (s *Service) CreateNotificationWithPriority(userID uint, refType string, refID uint, title, message, actionURL, priority string) error {
	notification := &domain.Notification{
		UserID:        userID,
		ReferenceType: refType,
		ReferenceID:   refID,
		Title:         title,
		Message:       message,
		ActionURL:     actionURL,
		IsRead:        false,
		Priority:      priority,
	}

	return s.repo.Create(notification)
}

// GetUserNotifications returns notifications for a user with optional filtering
func (s *Service) GetUserNotifications(userID uint, isRead *bool, page, limit int) ([]domain.Notification, int64, error) {
	filters := make(map[string]interface{})

	if isRead != nil {
		filters["is_read"] = *isRead
	}

	if page > 0 {
		filters["page"] = page
	}

	if limit > 0 {
		filters["limit"] = limit
	}

	notifications, err := s.repo.GetByUserID(userID, filters)
	if err != nil {
		return nil, 0, err
	}

	unreadCount, err := s.repo.GetUnreadCount(userID)
	if err != nil {
		return nil, 0, err
	}

	return notifications, unreadCount, nil
}

// MarkAsRead marks a single notification as read
func (s *Service) MarkAsRead(notificationID, userID uint) error {
	// Verify notification belongs to user
	notification, err := s.repo.GetByID(notificationID)
	if err != nil {
		return errors.New("notification not found")
	}

	if notification.UserID != userID {
		return errors.New("notification does not belong to user")
	}

	return s.repo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *Service) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *Service) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

// NotifyTeamInvitation sends a notification for team invitation
func (s *Service) NotifyTeamInvitation(userID uint, teamID uint, teamName string, inviterName string) error {
	return s.CreateNotification(
		userID,
		"team_invitation",
		teamID,
		"Team Invitation",
		inviterName+" invited you to join team '"+teamName+"'",
		"/teams/"+string(rune(teamID)),
	)
}

// NotifyProposalFeedback sends a notification when proposal receives feedback
func (s *Service) NotifyProposalFeedback(userID uint, proposalID uint, decision string) error {
	var title, message string
	switch decision {
	case "approve":
		title = "Proposal Approved"
		message = "Your proposal has been approved!"
	case "revise":
		title = "Revision Requested"
		message = "Your proposal requires revision. Please check the feedback."
	case "reject":
		title = "Proposal Rejected"
		message = "Unfortunately, your proposal has been rejected."
	default:
		title = "Proposal Feedback"
		message = "You have received feedback on your proposal."
	}

	return s.CreateNotificationWithPriority(
		userID,
		"proposal",
		proposalID,
		title,
		message,
		"/proposals/"+string(rune(proposalID)),
		"high",
	)
}

// NotifyProjectPublished sends a notification when a project is published
func (s *Service) NotifyProjectPublished(userID uint, projectID uint, projectTitle string) error {
	return s.CreateNotification(
		userID,
		"project",
		projectID,
		"Project Published",
		"Your project '"+projectTitle+"' has been published to the public archive!",
		"/projects/"+string(rune(projectID)),
	)
}
