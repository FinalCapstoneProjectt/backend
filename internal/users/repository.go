package users

import (
	"backend/internal/domain"

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
	SearchStudents(query string, departmentID uint) ([]domain.User, error)
	GetTeachers(departmentID uint) ([]domain.User, error)
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

func (r *repository) SearchStudents(query string, departmentID uint) ([]domain.User, error) {
	var users []domain.User
	db := r.db.Preload("University").Preload("Department").Where("role = ?", "student").Where("is_active = ?", true)

	if query != "" {
		searchQuery := "%" + query + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ? OR student_id ILIKE ?", searchQuery, searchQuery, searchQuery)
	}

	if departmentID > 0 {
		db = db.Where("department_id = ?", departmentID)
	}

	err := db.Limit(20).Find(&users).Error
	return users, err
}

func (r *repository) GetTeachers(departmentID uint) ([]domain.User, error) {
	var users []domain.User
	db := r.db.Preload("University").Preload("Department").Where("role = ?", "teacher").Where("is_active = ?", true)

	if departmentID > 0 {
		db = db.Where("department_id = ?", departmentID)
	}

	err := db.Find(&users).Error
	return users, err
}
