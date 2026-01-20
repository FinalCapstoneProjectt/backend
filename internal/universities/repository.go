package universities

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(university *domain.University) error
	GetByID(id uint) (*domain.University, error)
	GetAll() ([]domain.University, error)
	Update(university *domain.University) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(university *domain.University) error {
	return r.db.Create(university).Error
}

func (r *repository) GetByID(id uint) (*domain.University, error) {
	var university domain.University
	err := r.db.First(&university, id).Error
	if err != nil {
		return nil, err
	}
	return &university, nil
}

func (r *repository) GetAll() ([]domain.University, error) {
	var universities []domain.University
	err := r.db.Find(&universities).Error
	return universities, err
}

func (r *repository) Update(university *domain.University) error {
	return r.db.Save(university).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.University{}, id).Error
}
