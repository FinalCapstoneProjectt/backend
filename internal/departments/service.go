package departments

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

type CreateDepartmentRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code"`
	UniversityID uint   `json:"university_id" binding:"required"`
}

type UpdateDepartmentRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (s *Service) CreateDepartment(req CreateDepartmentRequest) (*domain.Department, error) {
	if req.Name == "" {
		return nil, errors.New("department name is required")
	}
	if req.UniversityID == 0 {
		return nil, errors.New("university ID is required")
	}

	department := &domain.Department{
		Name:         req.Name,
		Code:         req.Code,
		UniversityID: req.UniversityID,
	}

	err := s.repo.Create(department)
	if err != nil {
		return nil, err
	}

	// Fetch with university preloaded
	return s.repo.GetByID(department.ID)
}

func (s *Service) GetDepartment(id uint) (*domain.Department, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAllDepartments() ([]domain.Department, error) {
	return s.repo.GetAll()
}

func (s *Service) GetDepartmentsByUniversity(universityID uint) ([]domain.Department, error) {
	return s.repo.GetByUniversityID(universityID)
}

func (s *Service) UpdateDepartment(id uint, req UpdateDepartmentRequest) (*domain.Department, error) {
	department, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("department not found")
	}

	if req.Name != "" {
		department.Name = req.Name
	}
	if req.Code != "" {
		department.Code = req.Code
	}

	err = s.repo.Update(department)
	if err != nil {
		return nil, err
	}

	return department, nil
}

func (s *Service) DeleteDepartment(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("department not found")
	}

	return s.repo.Delete(id)
}
