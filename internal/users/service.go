package users

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

type CreateTeacherRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	UniversityID uint   `json:"university_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
}

type CreateStudentRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	StudentID    string `json:"student_id" binding:"required"`
	UniversityID uint   `json:"university_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type AssignDepartmentRequest struct {
	DepartmentID uint `json:"department_id" binding:"required"`
}

type UserResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         enums.Role `json:"role"`
	StudentID    string     `json:"student_id,omitempty"`
	UniversityID uint       `json:"university_id"`
	DepartmentID uint       `json:"department_id"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    string     `json:"created_at"`
}

func (s *Service) CreateTeacher(req CreateTeacherRequest) (*domain.User, error) {
	// Check if email already exists
	existing, _ := s.repo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Role:         enums.RoleTeacher,
		UniversityID: req.UniversityID,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) CreateStudent(req CreateStudentRequest) (*domain.User, error) {
	// Check if email already exists
	existing, _ := s.repo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Role:         enums.RoleStudent,
		StudentID:    req.StudentID,
		UniversityID: req.UniversityID,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) GetUser(id uint) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAllUsers(role string, departmentID uint, universityID uint, isActive *bool) ([]domain.User, error) {
	filters := make(map[string]interface{})

	if role != "" {
		filters["role"] = role
	}
	if departmentID > 0 {
		filters["department_id"] = departmentID
	}
	if universityID > 0 {
		filters["university_id"] = universityID
	}
	if isActive != nil {
		filters["is_active"] = *isActive
	}

	return s.repo.GetAll(filters)
}

func (s *Service) UpdateUserStatus(id uint, isActive bool) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.UpdateStatus(id, isActive)
}

func (s *Service) AssignDepartment(userID uint, departmentID uint) error {
	_, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.AssignDepartment(userID, departmentID)
}

func (s *Service) DeleteUser(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.Delete(id)
}

func (s *Service) SearchStudents(query string, departmentID uint) ([]domain.User, error) {
	return s.repo.SearchStudents(query, departmentID)
}

func (s *Service) GetTeachers(departmentID uint) ([]domain.User, error) {
	return s.repo.GetTeachers(departmentID)
}
