package universities

import (
	"backend/internal/domain"
	"errors"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

type CreateUniversityRequest struct {
	Name             string `json:"name" binding:"required"`
	AcademicYear     string `json:"academic_year"`
	ProjectPeriod    string `json:"project_period"`
	VisibilityRule   string `json:"visibility_rule"`
	AICheckerEnabled bool   `json:"ai_checker_enabled"`
}

type UpdateUniversityRequest struct {
	Name             string `json:"name"`
	AcademicYear     string `json:"academic_year"`
	ProjectPeriod    string `json:"project_period"`
	VisibilityRule   string `json:"visibility_rule"`
	AICheckerEnabled *bool  `json:"ai_checker_enabled"`
}

func (s *Service) CreateUniversity(req CreateUniversityRequest) (*domain.University, error) {
	if req.Name == "" {
		return nil, errors.New("university name is required")
	}

	university := &domain.University{
		Name:             req.Name,
		AcademicYear:     req.AcademicYear,
		ProjectPeriod:    req.ProjectPeriod,
		VisibilityRule:   req.VisibilityRule,
		AICheckerEnabled: req.AICheckerEnabled,
	}

	if university.VisibilityRule == "" {
		university.VisibilityRule = "private"
	}

	err := s.repo.Create(university)
	if err != nil {
		return nil, err
	}

	return university, nil
}

func (s *Service) GetUniversity(id uint) (*domain.University, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAllUniversities() ([]domain.University, error) {
	return s.repo.GetAll()
}

func (s *Service) UpdateUniversity(id uint, req UpdateUniversityRequest) (*domain.University, error) {
	university, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("university not found")
	}

	if req.Name != "" {
		university.Name = req.Name
	}
	if req.AcademicYear != "" {
		university.AcademicYear = req.AcademicYear
	}
	if req.ProjectPeriod != "" {
		university.ProjectPeriod = req.ProjectPeriod
	}
	if req.VisibilityRule != "" {
		university.VisibilityRule = req.VisibilityRule
	}
	if req.AICheckerEnabled != nil {
		university.AICheckerEnabled = *req.AICheckerEnabled
	}

	err = s.repo.Update(university)
	if err != nil {
		return nil, err
	}

	return university, nil
}

func (s *Service) DeleteUniversity(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("university not found")
	}

	return s.repo.Delete(id)
}
