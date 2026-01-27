package reviews

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

// Repository defines the interface for project review data access
type Repository interface {
	Create(review *domain.ProjectReview) error
	GetByProjectID(projectID uint) ([]domain.ProjectReview, error)
	GetByUserAndProject(userID, projectID uint) (*domain.ProjectReview, error)
	GetAverageRating(projectID uint) (float64, error)
	Update(review *domain.ProjectReview) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new review repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(review *domain.ProjectReview) error {
	return r.db.Create(review).Error
}

func (r *repository) GetByProjectID(projectID uint) ([]domain.ProjectReview, error) {
	var reviews []domain.ProjectReview
	err := r.db.Where("project_id = ?", projectID).
		Preload("User").
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *repository) GetByUserAndProject(userID, projectID uint) (*domain.ProjectReview, error) {
	var review domain.ProjectReview
	err := r.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *repository) GetAverageRating(projectID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&domain.ProjectReview{}).
		Where("project_id = ?", projectID).
		Preload("User").
		Select("COALESCE(AVG(rate), 0)").
		Scan(&avg).Error
	return avg, err
}

func (r *repository) Update(review *domain.ProjectReview) error {
	return r.db.Save(review).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.ProjectReview{}, id).Error
}
