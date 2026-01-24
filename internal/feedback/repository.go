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
	GetDB() *gorm.DB
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
func (r *repository) GetDB() *gorm.DB {
	return r.db
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

func (r *repository) GetPendingProposalsForReviewer(advisorID uint) ([]domain.Proposal, error) {
	var proposals []domain.Proposal
	// 👈 FIX: Look at proposals.advisor_id and deep preload for the UI
	err := r.db.
		Preload("Team.Members.User").
		Preload("Team.Department").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC")
		}).
		Where("advisor_id = ?", advisorID). // 👈 Proposal's assigned advisor
		Where("status IN ?", []string{"submitted", "under_review", "revision_required", "approved", "rejected"}).
		Find(&proposals).Error

	return proposals, err
}