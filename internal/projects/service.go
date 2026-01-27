package projects

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
)

type Service struct {
	repo         Repository
	proposalRepo ProposalRepository
}

type ProposalRepository interface {
	GetByID(id uint) (*domain.Proposal, error)
}

func NewService(repo Repository, proposalRepo ProposalRepository) *Service {
	return &Service{
		repo:         repo,
		proposalRepo: proposalRepo,
	}
}

type CreateProjectRequest struct {
	ProposalID uint   `json:"proposal_id"`
	Summary    string `json:"summary"`
	Keywords   string `json:"keywords"`
}

type UpdateProjectRequest struct {
	Summary  string `json:"summary"`
	Visibility string `json:"visibility"`
}

func (s *Service) CreateProject(req CreateProjectRequest, userID uint) (*domain.Project, error) {
	// 1. Verify proposal exists and is approved
	proposal, err := s.proposalRepo.GetByID(req.ProposalID)
	if err != nil {
		return nil, errors.New("proposal not found")
	}

	if proposal.Status != enums.ProposalStatusApproved {
		return nil, errors.New("only approved proposals can become projects")
	}

	// 2. Check if project already exists for this proposal
	existing, _ := s.repo.GetByProposalID(req.ProposalID)
	if existing != nil {
		return nil, errors.New("project already exists for this proposal")
	}

	// 3. Get team info for department
	var teamID uint
	var departmentID uint
	if proposal.TeamID != nil {
		teamID = *proposal.TeamID
		if proposal.Team != nil {
			departmentID = proposal.Team.DepartmentID
		}
	}

	// 4. Create project
	project := &domain.Project{
		ProposalID:   req.ProposalID,
		TeamID:       teamID,
		DepartmentID: departmentID,
		Summary:      req.Summary,
		ApprovedBy:   userID,
		Visibility:   "private",
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	return s.repo.GetByID(project.ID)
}

func (s *Service) GetProject(id uint) (*domain.Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Increment view count
	_ = s.repo.IncrementViewCount(id)

	return project, nil
}

func (s *Service) GetProjects(filters map[string]interface{}) ([]domain.Project, error) {
	return s.repo.GetAll(filters)
}

func (s *Service) UpdateProject(id uint, req UpdateProjectRequest, userID uint, role enums.Role) (*domain.Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// Permission Logic (Allow Creator, Advisor, or Admin)
	isCreator := project.Team.CreatedBy == userID
	isAdvisor := project.Proposal.AdvisorID != nil && *project.Proposal.AdvisorID == userID
	isAdmin := role == enums.RoleAdmin

	if !isCreator && !isAdvisor && !isAdmin {
		return nil, errors.New("unauthorized: you cannot update this project")
	}

	// 🚀 Apply updates ONLY if they are provided in the JSON
	if req.Summary != "" {
		project.Summary = req.Summary
	}
	if req.Visibility != "" {
		project.Visibility = req.Visibility
	}

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) PublishProject(id uint, userID uint, role enums.Role) error {
		project, err := s.repo.GetByID(id)
	if err != nil { return err }

	// 🔒 FIX: Allow Creator OR Advisor OR Admin
	isCreator := project.Team.CreatedBy == userID
	isAdvisor := project.Proposal.AdvisorID != nil && *project.Proposal.AdvisorID == userID
	isAdmin := role == enums.RoleAdmin

	if !isCreator && !isAdvisor && !isAdmin {
		return errors.New("unauthorized: only team leader or assigned advisor can publish")
	}

	return s.repo.UpdateVisibility(id, "public")
}
