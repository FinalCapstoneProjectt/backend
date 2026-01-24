package documentations

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type Repository interface {
	Create(doc *domain.ProjectDocumentation) error
	GetByID(id uint) (*domain.ProjectDocumentation, error)
	GetByProjectID(projectID uint) ([]domain.ProjectDocumentation, error)
	GetByType(projectID uint, docType string) (*domain.ProjectDocumentation, error)
	Update(doc *domain.ProjectDocumentation) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(doc *domain.ProjectDocumentation) error { return r.db.Create(doc).Error }

func (r *repository) GetByID(id uint) (*domain.ProjectDocumentation, error) {
	var doc domain.ProjectDocumentation
	err := r.db.First(&doc, id).Error
	return &doc, err
}

func (r *repository) GetByProjectID(projectID uint) ([]domain.ProjectDocumentation, error) {
	var docs []domain.ProjectDocumentation
	err := r.db.Where("project_id = ?", projectID).Find(&docs).Error
	return docs, err
}

func (r *repository) GetByType(projectID uint, docType string) (*domain.ProjectDocumentation, error) {
	var doc domain.ProjectDocumentation
	err := r.db.Where("project_id = ? AND document_type = ?", projectID, docType).First(&doc).Error
	return &doc, err
}

func (r *repository) Update(doc *domain.ProjectDocumentation) error { return r.db.Save(doc).Error }

func (r *repository) Delete(id uint) error { return r.db.Delete(&domain.ProjectDocumentation{}, id).Error }

func (r *repository) IncrementViewCount(id uint) error {
    // ⚠️ Match the field "view_count" added in Step 1
	return r.db.Model(&domain.Project{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + ?", 1)).Error
}