package reviews

import (
	"backend/internal/domain"
	"errors"
	"time"
)

// Service handles project review business logic
type Service struct {
	repo        Repository
	projectRepo ProjectRepository
}

// ProjectRepository interface for accessing project data
type ProjectRepository interface {
	GetByID(id uint) (*domain.Project, error)
}

// NewService creates a new review service
func NewService(repo Repository, projectRepo ProjectRepository) *Service {
	return &Service{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// CreateReview creates a new review for a project
func (s *Service) CreateReview(userID, projectID uint, rating int, comment string) (*domain.ProjectReview, float64, error) {
	// Verify project exists and is public
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, 0, errors.New("project not found")
	}

	if project.Visibility != "public" {
		return nil, 0, errors.New("can only review public projects")
	}

	// Check if user already reviewed this project
	existing, _ := s.repo.GetByUserAndProject(userID, projectID)
	if existing != nil {
		return nil, 0, errors.New("you have already reviewed this project")
	}

	// Validate rating
	if rating < 1 || rating > 5 {
		return nil, 0, errors.New("rating must be between 1 and 5")
	}

	// Create review
	review := &domain.ProjectReview{
		ProjectID: projectID,
		UserID:    userID,
		Rate:      rating,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(review); err != nil {
		return nil, 0, err
	}

	// Get updated average rating
	avgRating, err := s.repo.GetAverageRating(projectID)
	if err != nil {
		avgRating = float64(rating)
	}

	return review, avgRating, nil
}

// GetProjectReviews returns all reviews for a project with average rating
func (s *Service) GetProjectReviews(projectID uint) ([]domain.ProjectReview, float64, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, 0, errors.New("project not found")
	}

	reviews, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, 0, err
	}

	avgRating, err := s.repo.GetAverageRating(projectID)
	if err != nil {
		avgRating = 0
	}

	return reviews, avgRating, nil
}

// GetAverageRating returns the average rating for a project
func (s *Service) GetAverageRating(projectID uint) (float64, error) {
	return s.repo.GetAverageRating(projectID)
}

// UpdateReview updates an existing review (only by the creator)
func (s *Service) UpdateReview(reviewID, userID uint, rating int, comment string) (*domain.ProjectReview, error) {
	review, err := s.repo.GetByUserAndProject(userID, reviewID)
	if err != nil {
		return nil, errors.New("review not found or not owned by user")
	}

	if rating >= 1 && rating <= 5 {
		review.Rate = rating
	}

	if comment != "" {
		review.Comment = comment
	}

	if err := s.repo.Update(review); err != nil {
		return nil, err
	}

	return review, nil
}

// DeleteReview deletes a review (only by the creator or admin)
func (s *Service) DeleteReview(reviewID, userID uint, isAdmin bool) error {
	_, err := s.repo.GetByUserAndProject(userID, reviewID)
	if err != nil && !isAdmin {
		return errors.New("review not found or not owned by user")
	}

	return s.repo.Delete(reviewID)
}
