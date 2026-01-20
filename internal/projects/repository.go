package projects

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(project *domain.Project) error
	GetByID(id uint) (*domain.Project, error)
	GetByProposalID(proposalID uint) (*domain.Project, error)
	GetAll(filters map[string]interface{}) ([]domain.Project, error)
	Update(project *domain.Project) error
	UpdateVisibility(id uint, visibility string) error
	IncrementViewCount(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(project *domain.Project) error {
	return r.db.Create(project).Error
}

func (r *repository) GetByID(id uint) (*domain.Project, error) {
	var project domain.Project
	err := r.db.Preload("Proposal").
		Preload("Team").
		Preload("Team.Members").
		Preload("Department").
		Preload("Approver").
		First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *repository) GetByProposalID(proposalID uint) (*domain.Project, error) {
	var project domain.Project
	err := r.db.Where("proposal_id = ?", proposalID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *repository) GetAll(filters map[string]interface{}) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.Preload("Team").
		Preload("Department").
		Preload("Proposal")

	if visibility, ok := filters["visibility"]; ok {
		query = query.Where("visibility = ?", visibility)
	}
	if departmentID, ok := filters["department_id"]; ok {
		query = query.Where("department_id = ?", departmentID)
	}
	if teamID, ok := filters["team_id"]; ok {
		query = query.Where("team_id = ?", teamID)
	}

	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

func (r *repository) Update(project *domain.Project) error {
	return r.db.Save(project).Error
}

func (r *repository) UpdateVisibility(id uint, visibility string) error {
	return r.db.Model(&domain.Project{}).
		Where("id = ?", id).
		Update("visibility", visibility).Error
}

func (r *repository) IncrementViewCount(id uint) error {
	return r.db.Model(&domain.Project{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + ?", 1)).Error
}
