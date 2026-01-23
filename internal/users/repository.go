package users

import (
	"backend/internal/domain"
	"backend/pkg/enums" // Make sure to import this!

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetAll(filters map[string]interface{}) ([]domain.User, error)
	Update(user *domain.User) error
	UpdateStatus(id uint, isActive bool) error
	AssignDepartment(userID uint, departmentID uint) error
	Delete(id uint) error
	FindPeers(departmentID uint, universityID uint, excludeUserID uint) ([]domain.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *repository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("University").Preload("Department").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetAll(filters map[string]interface{}) ([]domain.User, error) {
	var users []domain.User
	query := r.db.Preload("University").Preload("Department")

	if role, ok := filters["role"]; ok {
		query = query.Where("role = ?", role)
	}
	if departmentID, ok := filters["department_id"]; ok {
		query = query.Where("department_id = ?", departmentID)
	}
	if universityID, ok := filters["university_id"]; ok {
		query = query.Where("university_id = ?", universityID)
	}
	if isActive, ok := filters["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}

	err := query.Find(&users).Error
	return users, err
}

func (r *repository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *repository) UpdateStatus(id uint, isActive bool) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *repository) AssignDepartment(userID uint, departmentID uint) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("department_id", departmentID).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *repository) FindPeers(departmentID uint, universityID uint, excludeUserID uint) ([]domain.User, error) {
	var users []domain.User
	// Fetch users who are:
	// 1. In the same University
	// 2. In the same Department
	// 3. Are Students
	// 4. Are NOT the requester (excludeUserID)
	err := r.db.Where("university_id = ? AND department_id = ? AND role = ? AND id != ?", 
		universityID, departmentID, enums.RoleStudent, excludeUserID).
		Find(&users).Error
	return users, err
}