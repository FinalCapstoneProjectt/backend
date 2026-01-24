package feedback

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"

	"gorm.io/gorm" 
)

type Service struct {
	repo         Repository
	proposalRepo ProposalRepository
}

// Ensure this matches your proposals.Repository interface
type ProposalRepository interface {
	GetByID(id uint) (*domain.Proposal, error)
	Update(proposal *domain.Proposal) error
}

func NewService(repo Repository, proposalRepo ProposalRepository) *Service {
	return &Service{repo: repo, proposalRepo: proposalRepo}
}

type CreateFeedbackRequest struct {
	ProposalID        uint   `json:"proposal_id" binding:"required"`
	ProposalVersionID uint   `json:"proposal_version_id" binding:"required"`
	Decision          string `json:"decision" binding:"required"` // approve, revise, reject
	Comment           string `json:"comment" binding:"required"`
}
func (s *Service) CreateFeedback(req CreateFeedbackRequest, reviewerID uint) (*domain.Feedback, error) {
	// 1. Get proposal
	proposal, err := s.proposalRepo.GetByID(req.ProposalID)
	if err != nil { return nil, errors.New("proposal not found") }

	// 2. Security Check
	if proposal.AdvisorID == nil || *proposal.AdvisorID != reviewerID {
		return nil, errors.New("only the assigned advisor can review this proposal")
	}

	feedback := &domain.Feedback{
		ProposalID:        req.ProposalID,
		ProposalVersionID: req.ProposalVersionID,
		ReviewerID:        reviewerID,
		Decision:          domain.FeedbackDecision(req.Decision),
		Comment:           req.Comment,
	}

	// 3. Handle Decision
	if req.Decision == "approve" {
		// 🚨 SAFETY CHECKS (Prevents Panic)
		if proposal.TeamID == nil {
			return nil, errors.New("cannot approve: proposal is not linked to a team")
		}
		if proposal.Team == nil {
			return nil, errors.New("cannot approve: team data failed to load")
		}

		var versionAbstract string
		for _, v := range proposal.Versions {
			if v.ID == req.ProposalVersionID {
				versionAbstract = v.Abstract
			}
		}

		// Run Transaction
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(feedback).Error; err != nil { return err }

			// Update Status
			if err := tx.Model(&domain.Proposal{}).Where("id = ?", proposal.ID).Update("status", enums.ProposalStatusApproved).Error; err != nil { return err }

			// Mark version approved
			if err := tx.Model(&domain.ProposalVersion{}).Where("id = ?", req.ProposalVersionID).Update("is_approved", true).Error; err != nil { return err }

			// Create Project
			project := &domain.Project{
				ProposalID:   proposal.ID,
				TeamID:       *proposal.TeamID, // Now safe to dereference
				DepartmentID: proposal.Team.DepartmentID, // Now safe
				Summary:      versionAbstract,
				ApprovedBy:   reviewerID,
				Visibility:   "private",
			}
			return tx.Create(project).Error
		})
		if err != nil { return nil, err }

	} else {
		// Logic for Revise/Reject
		if err := s.repo.Create(feedback); err != nil { return nil, err }
		
		newStatus := enums.ProposalStatusRejected
		if req.Decision == "revise" {
			newStatus = enums.ProposalStatusRevisionRequired
		}
		
		if err := s.repo.GetDB().Model(&domain.Proposal{}).Where("id = ?", req.ProposalID).Update("status", newStatus).Error; err != nil { return nil, err }
	}

	return feedback, nil
}

// Helper to update status
func txUpdateStatus(db *gorm.DB, id uint, status enums.ProposalStatus) error {
	return db.Model(&domain.Proposal{}).Where("id = ?", id).Update("status", status).Error
}

func (s *Service) GetProposalFeedback(proposalID uint, userID uint) ([]domain.Feedback, error) {
	// Logic: Fetch all feedback for this proposal
	return s.repo.GetByProposalID(proposalID)
}

func (s *Service) GetPendingProposals(reviewerID uint) ([]domain.Proposal, error) {
	return s.repo.GetPendingProposalsForReviewer(reviewerID)
}

func (s *Service) GetFeedbackByID(id uint) (*domain.Feedback, error) {
	return s.repo.GetByID(id)
}