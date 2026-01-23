package proposals

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"

	"gorm.io/gorm"
)

type Service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(r Repository, db *gorm.DB) *Service {
	return &Service{repo: r, db: db}
}

// DTO for Service Input
type ProposalInput struct {
	TeamID           *uint 
	Title            string
	Abstract         string
	ProblemStatement string
	Objectives       string
	Methodology      string
	Timeline         string
	ExpectedOutcomes string
}

// 1. Create New Draft (Creates Proposal + Version 1)
func (s *Service) CreateDraft(input ProposalInput, userID uint) (*domain.Proposal, error) {
	var proposal domain.Proposal
	
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Parent (Status: Draft)
		proposal = domain.Proposal{
			TeamID:    input.TeamID,
			Status:    enums.ProposalStatusDraft,
			AdvisorID: nil,
		}
		if err := tx.Create(&proposal).Error; err != nil { return err }

		// 2. Create Version 1
		version := domain.ProposalVersion{
			ProposalID:       proposal.ID,
			CreatedBy: userID,
			VersionNumber:    1,
			Title:            input.Title,
			Abstract:         input.Abstract,
			ProblemStatement: input.ProblemStatement,
			Objectives:       input.Objectives,
			Methodology:      input.Methodology,
			ExpectedTimeline: input.Timeline,
			ExpectedOutcomes: input.ExpectedOutcomes,
			FileURL:         nil,
			FileHash:      "",
    		FileSizeBytes: 0,
		}
		return tx.Create(&version).Error
	})
	return &proposal, err
}

// 2. Update Proposal (Edit Draft OR Create Revision)
func (s *Service) UpdateProposal(proposalID uint, input ProposalInput, userID uint) (*domain.Proposal, error) {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil { return nil, err }

	// Rule: Check if status allows editing (Draft, Rejected, RevisionRequired)
	if !CanEdit(proposal.Status) {
		return nil, errors.New("proposal is locked and cannot be edited")
	}

	// Scenario A: It is a DRAFT -> Overwrite Version 1
	if proposal.Status == enums.ProposalStatusDraft {
		return s.overwriteDraftVersion(proposal, input)
	}

	// Scenario B: It is REJECTED or REVISION -> Create NEW Version (History)
	return s.createNewVersion(proposal, input, userID)
}

// Internal: Overwrites Version 1 directly
func (s *Service) overwriteDraftVersion(p *domain.Proposal, input ProposalInput) (*domain.Proposal, error) {
	version, err := s.repo.GetFirstVersion(p.ID)
	if err != nil { return nil, err }

	// Update Fields
	version.Title = input.Title
	version.Abstract = input.Abstract
	version.ProblemStatement = input.ProblemStatement
	version.Objectives = input.Objectives
	version.Methodology = input.Methodology
	version.ExpectedTimeline = input.Timeline

	// Update Team if changed
	if input.TeamID != nil {
		p.TeamID = input.TeamID
		if err := s.repo.Update(p); err != nil { return nil, err }
	}

	if err := s.db.Save(version).Error; err != nil { return nil, err }
	return p, nil
}

// Internal: Creates V+1
func (s *Service) createNewVersion(p *domain.Proposal, input ProposalInput, userID uint) (*domain.Proposal, error) {
	lastVer, err := s.repo.GetLatestVersion(p.ID)
	if err != nil { return nil, err }

	newVer := domain.ProposalVersion{
		ProposalID:       p.ID,
		CreatedBy: userID,
		VersionNumber:    lastVer.VersionNumber + 1,
		Title:            input.Title,
		Abstract:         input.Abstract,
		ProblemStatement: input.ProblemStatement,
		Objectives:       input.Objectives,
		Methodology:      input.Methodology,
		ExpectedTimeline: input.Timeline,
		ExpectedOutcomes: input.ExpectedOutcomes, 
		FileHash:      "",
   		FileSizeBytes: 0,

		FileURL:         nil,
	}

	if err := s.repo.CreateVersion(&newVer); err != nil { return nil, err }
	return p, nil
}

// 3. Submit Proposal
func (s *Service) SubmitProposal(proposalID uint, teamID uint, userID uint) error {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil { return err }

	// Rule: Valid Status?
	if !CanSubmit(proposal.Status) {
		return errors.New("proposal cannot be submitted in current state")
	}

	// Rule: Fetch Team & Check Finalized
	var team domain.Team
	if err := s.db.Preload("Members").First(&team, teamID).Error; err != nil {
		return errors.New("team not found")
	}

	if !team.IsFinalized {
		return errors.New("selected team is not finalized")
	}

	// Rule: Is User Leader?
	isLeader := false
	for _, m := range team.Members {
		if m.UserID == userID && m.Role == "leader" {
			isLeader = true; break
		}
	}
	if !isLeader {
		return errors.New("only team leader can submit")
	}

	// Update Status to Submitted
	proposal.TeamID = &teamID
	proposal.Status = enums.ProposalStatusSubmitted
	
	return s.repo.Update(proposal)
}

// Getters
func (s *Service) GetProposals(status string, deptID uint) ([]domain.Proposal, error) {
	filters := make(map[string]interface{})
	if status != "" { filters["status"] = status }
	if deptID != 0 { filters["department_id"] = deptID }
	return s.repo.GetAll(filters)
}

func (s *Service) GetProposal(id uint) (*domain.Proposal, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetVersions(id uint) ([]domain.ProposalVersion, error) {
	return s.repo.GetVersionsByProposalID(id)
}

func (s *Service) DeleteProposal(id uint) error {
	proposal, err := s.repo.GetByID(id)
	if err != nil { return err }

	if proposal.Status != enums.ProposalStatusDraft {
		return errors.New("only draft proposals can be deleted")
	}
	return s.repo.Delete(id)
}