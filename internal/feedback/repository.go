package feedback

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(feedback *domain.Feedback) error
	GetByProposalID(proposalID uint) ([]domain.Feedback, error)
	GetByID(id uint) (*domain.Feedback, error)
	GetPendingProposalsForReviewer(reviewerID uint) ([]domain.Proposal, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(feedback *domain.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *repository) GetByProposalID(proposalID uint) ([]domain.Feedback, error) {
	var feedbacks []domain.Feedback
	err := r.db.Preload("Reviewer").
		Where("proposal_id = ?", proposalID).
		Order("created_at DESC").
		Find(&feedbacks).Error
	return feedbacks, err
}

func (r *repository) GetByID(id uint) (*domain.Feedback, error) {
	var feedback domain.Feedback
	err := r.db.Preload("Reviewer").
		Preload("Proposal").
		Preload("ProposalVersion").
		First(&feedback, id).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *repository) GetPendingProposalsForReviewer(reviewerID uint) ([]domain.Proposal, error) {
	var proposals []domain.Proposal

	// Get proposals where:
	// 1. Status is submitted or under_review
	// 2. Team's advisor is the reviewer
	err := r.db.Preload("Team").
		Preload("Team.Creator").
		Preload("Team.Department").
		Preload("CurrentVersion").
		Joins("JOIN teams ON proposals.team_id = teams.id").
		Where("teams.advisor_id = ?", reviewerID).
		Where("proposals.status IN ?", []string{"submitted", "under_review"}).
		Find(&proposals).Error

	return proposals, err
}
