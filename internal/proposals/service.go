package proposals

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

type CreateProposalRequest struct {
	TeamID uint `json:"team_id" binding:"required"`
}

type CreateVersionRequest struct {
	Title            string `json:"title" binding:"required"`
	Objectives       string `json:"objectives" binding:"required"`
	Methodology      string `json:"methodology" binding:"required"`
	ExpectedOutcomes string `json:"expected_outcomes" binding:"required"`
	FileURL          string `json:"file_url" binding:"required"`
	FileHash         string `json:"file_hash" binding:"required"`
	FileSizeBytes    int64  `json:"file_size_bytes" binding:"required"`
}

type UpdateVersionRequest struct {
	Title            string `json:"title"`
	Objectives       string `json:"objectives"`
	Methodology      string `json:"methodology"`
	ExpectedOutcomes string `json:"expected_outcomes"`
	FileURL          string `json:"file_url"`
	FileHash         string `json:"file_hash"`
	FileSizeBytes    int64  `json:"file_size_bytes"`
}

func (s *Service) CreateProposal(req CreateProposalRequest) (*domain.Proposal, error) {
	// Check if team already has a proposal
	existing, _ := s.repo.GetByTeamID(req.TeamID)
	if existing != nil {
		return nil, errors.New("team already has a proposal")
	}

	proposal := &domain.Proposal{
		TeamID: req.TeamID,
		Status: enums.ProposalStatusDraft,
	}

	err := s.repo.Create(proposal)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(proposal.ID)
}

func (s *Service) GetProposal(id uint) (*domain.Proposal, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetProposals(status string, departmentID uint) ([]domain.Proposal, error) {
	filters := make(map[string]interface{})

	if status != "" {
		filters["status"] = status
	}
	if departmentID > 0 {
		filters["department_id"] = departmentID
	}

	return s.repo.GetAll(filters)
}

func (s *Service) GetProposalByTeam(teamID uint) (*domain.Proposal, error) {
	return s.repo.GetByTeamID(teamID)
}

func (s *Service) CreateVersion(proposalID uint, req CreateVersionRequest, creatorID uint) (*domain.ProposalVersion, error) {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil {
		return nil, errors.New("proposal not found")
	}

	// Only allow creating versions for draft or revision required proposals
	if proposal.Status != enums.ProposalStatusDraft && proposal.Status != enums.ProposalStatusRevisionRequired {
		return nil, errors.New("cannot create version for this proposal status")
	}

	// Get existing versions to determine version number
	versions, _ := s.repo.GetVersionsByProposalID(proposalID)
	versionNumber := len(versions) + 1

	version := &domain.ProposalVersion{
		ProposalID:       proposalID,
		Title:            req.Title,
		Objectives:       req.Objectives,
		Methodology:      req.Methodology,
		ExpectedOutcomes: req.ExpectedOutcomes,
		FileURL:          req.FileURL,
		FileHash:         req.FileHash,
		FileSizeBytes:    req.FileSizeBytes,
		VersionNumber:    versionNumber,
		CreatedBy:        creatorID,
	}

	err = s.repo.CreateVersion(version)
	if err != nil {
		return nil, err
	}

	// Update proposal's current version
	proposal.CurrentVersionID = &version.ID
	s.repo.Update(proposal)

	return s.repo.GetVersionByID(version.ID)
}

func (s *Service) GetVersions(proposalID uint) ([]domain.ProposalVersion, error) {
	return s.repo.GetVersionsByProposalID(proposalID)
}

func (s *Service) SubmitProposal(proposalID uint) error {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil {
		return errors.New("proposal not found")
	}

	// Must have at least one version
	if proposal.CurrentVersionID == nil {
		return errors.New("proposal must have at least one version before submission")
	}

	// Can only submit from draft or revision required status
	if proposal.Status != enums.ProposalStatusDraft && proposal.Status != enums.ProposalStatusRevisionRequired {
		return errors.New("cannot submit proposal in current status")
	}

	now := time.Now()
	proposal.Status = enums.ProposalStatusSubmitted
	proposal.SubmittedAt = &now
	proposal.SubmissionCount++

	return s.repo.Update(proposal)
}

func (s *Service) DeleteProposal(id uint) error {
	proposal, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("proposal not found")
	}

	// Can only delete drafts
	if proposal.Status != enums.ProposalStatusDraft {
		return errors.New("can only delete draft proposals")
	}

	return s.repo.Delete(id)
}

func (s *Service) UpdateProposalStatus(id uint, status enums.ProposalStatus) error {
	return s.repo.UpdateProposalStatus(id, status)
}
