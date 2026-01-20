package feedback

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
)

type Service struct {
	repo         Repository
	proposalRepo ProposalRepository
}

// ProposalRepository interface for proposal state changes
type ProposalRepository interface {
	GetByID(id uint) (*domain.Proposal, error)
	UpdateProposalStatus(id uint, status enums.ProposalStatus) error
}

func NewService(repo Repository, proposalRepo ProposalRepository) *Service {
	return &Service{
		repo:         repo,
		proposalRepo: proposalRepo,
	}
}

type CreateFeedbackRequest struct {
	ProposalID        uint                    `json:"proposal_id" binding:"required"`
	ProposalVersionID uint                    `json:"proposal_version_id" binding:"required"`
	Decision          domain.FeedbackDecision `json:"decision" binding:"required"`
	Comment           string                  `json:"comment" binding:"required,min=20"`
}

func (s *Service) CreateFeedback(req CreateFeedbackRequest, reviewerID uint) (*domain.Feedback, error) {
	// 1. Get proposal to validate
	proposal, err := s.proposalRepo.GetByID(req.ProposalID)
	if err != nil {
		return nil, err
	}

	// 2. Validate proposal is in reviewable state
	if proposal.Status != enums.ProposalStatusSubmitted && proposal.Status != enums.ProposalStatusUnderReview {
		return nil, errors.New("proposal is not in a reviewable state")
	}

	// 3. Validate reviewer is the team's advisor
	if proposal.Team.AdvisorID != reviewerID {
		return nil, errors.New("only the assigned advisor can review this proposal")
	}

	// 4. Create feedback record
	feedback := &domain.Feedback{
		ProposalID:        req.ProposalID,
		ProposalVersionID: req.ProposalVersionID,
		ReviewerID:        reviewerID,
		Decision:          req.Decision,
		Comment:           req.Comment,
	}

	if err := s.repo.Create(feedback); err != nil {
		return nil, err
	}

	// 5. Update proposal status based on decision
	var newStatus enums.ProposalStatus
	switch req.Decision {
	case domain.FeedbackDecisionApprove:
		newStatus = enums.ProposalStatusApproved
	case domain.FeedbackDecisionRevise:
		newStatus = enums.ProposalStatusRevisionRequired
	case domain.FeedbackDecisionReject:
		newStatus = enums.ProposalStatusRejected
	}

	if err := s.proposalRepo.UpdateProposalStatus(req.ProposalID, newStatus); err != nil {
		return nil, err
	}

	return feedback, nil
}

func (s *Service) GetProposalFeedback(proposalID uint, userID uint) ([]domain.Feedback, error) {
	// Get the proposal to check permissions
	proposal, err := s.proposalRepo.GetByID(proposalID)
	if err != nil {
		return nil, err
	}

	// Check if user has access (team creator or advisor)
	if proposal.Team.CreatedBy != userID && proposal.Team.AdvisorID != userID {
		return nil, errors.New("you don't have permission to view this feedback")
	}

	return s.repo.GetByProposalID(proposalID)
}

func (s *Service) GetPendingProposals(reviewerID uint) ([]domain.Proposal, error) {
	return s.repo.GetPendingProposalsForReviewer(reviewerID)
}

func (s *Service) GetFeedbackByID(feedbackID uint) (*domain.Feedback, error) {
	return s.repo.GetByID(feedbackID)
}
