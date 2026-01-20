package departments

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(department *domain.Department) error
	GetByID(id uint) (*domain.Department, error)
	GetAll() ([]domain.Department, error)
	GetByUniversityID(universityID uint) ([]domain.Department, error)
	Update(department *domain.Department) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(department *domain.Department) error {
	return r.db.Create(department).Error
}

func (r *repository) GetByID(id uint) (*domain.Department, error) {
	var department domain.Department
	err := r.db.Preload("University").First(&department, id).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

func (r *repository) GetAll() ([]domain.Department, error) {
	var departments []domain.Department
	err := r.db.Preload("University").Find(&departments).Error
	return departments, err
}

func (r *repository) GetByUniversityID(universityID uint) ([]domain.Department, error) {
	var departments []domain.Department
	err := r.db.Preload("University").Where("university_id = ?", universityID).Find(&departments).Error
	return departments, err
}

func (r *repository) Update(department *domain.Department) error {
	return r.db.Save(department).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Department{}, id).Error
}
