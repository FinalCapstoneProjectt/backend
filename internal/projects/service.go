package projects

import (
	"backend/internal/domain"
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
	ProposalID uint   `json:"proposal_id" binding:"required"`
	Summary    string `json:"summary" binding:"required,min=50"`
	Keywords   string `json:"keywords" binding:"required"`
}

type UpdateProjectRequest struct {
	Summary  string `json:"summary" binding:"required,min=50"`
	Keywords string `json:"keywords" binding:"required"`
}

func (s *Service) CreateProject(req CreateProjectRequest, userID uint) (*domain.Project, error) {
	// // 1. Verify proposal exists and is approved
	// proposal, err := s.proposalRepo.GetByID(req.ProposalID)
	// if err != nil {
	// 	return nil, err
	// }

	// if proposal.Status != "approved" {
	// 	return nil, errors.New("only approved proposals can become projects")
	// }

	// // 2. Check if project already exists for this proposal
	// existing, _ := s.repo.GetByProposalID(req.ProposalID)
	// if existing != nil {
	// 	return nil, errors.New("project already exists for this proposal")
	// }

	// // 3. Create project
	// project := &domain.Project{
	// 	ProposalID:   req.ProposalID,
	// 	TeamID:       proposal.TeamID,
	// 	DepartmentID: proposal.Team.DepartmentID,
	// 	Summary:      req.Summary,
	// 	ApprovedBy:   userID,
	// 	Visibility:   "private",
	// }

	// if err := s.repo.Create(project); err != nil {
	// 	return nil, err
	// }

	// return s.repo.GetByID(project.ID)
	return nil, errors.New("not implemented")
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

func (s *Service) UpdateProject(id uint, req UpdateProjectRequest, userID uint) (*domain.Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if user is team creator
	if project.Team.CreatedBy != userID {
		return nil, errors.New("only team creator can update project")
	}

	project.Summary = req.Summary

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	return s.repo.GetByID(id)
}

func (s *Service) PublishProject(id uint, userID uint) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Check if user is team creator or admin
	if project.Team.CreatedBy != userID {
		return errors.New("only team creator can publish project")
	}

	if project.Visibility == "public" {
		return errors.New("project is already public")
	}

	return s.repo.UpdateVisibility(id, "public")
}
