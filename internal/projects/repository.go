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
	GetPublicProjects(filters map[string]interface{}) ([]domain.Project, int, error)
	Update(project *domain.Project) error
	UpdateVisibility(id uint, visibility string) error
	IncrementViewCount(id uint) error
	IncrementShareCount(id uint) (int, error)
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
	err := r.db.
		Preload("Proposal.Versions").
		Preload("Team.Members.User"). 
		Preload("Team.Department").
		First(&project, id).Error
	return &project, err
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
	query := r.db.
		Preload("Team.Members.User").
		Preload("Proposal.Advisor").
		Preload("Department"). // 👈 Now this works
		Preload("Proposal.Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC")
		})

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
	// ⚠️ Use Omit to prevent GORM from trying to re-save the Team or Proposal objects
	return r.db.Model(project).Omit("Team", "Proposal", "Department", "Approver").Updates(project).Error
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

func (r *repository) IncrementShareCount(id uint) (int, error) {
	err := r.db.Model(&domain.Project{}).
		Where("id = ?", id).
		Update("share_count", gorm.Expr("share_count + ?", 1)).Error
	if err != nil {
		return 0, err
	}
	
	var project domain.Project
	r.db.Select("share_count").First(&project, id)
	return project.ShareCount, nil
}

func (r *repository) GetPublicProjects(filters map[string]interface{}) ([]domain.Project, int, error) {
	var projects []domain.Project
	var total int64

	query := r.db.Model(&domain.Project{}).Where("visibility = ?", "public")

	// Apply filters
	if deptID, ok := filters["department_id"]; ok {
		query = query.Where("department_id = ?", deptID)
	}
	if year, ok := filters["year"]; ok {
		query = query.Where("EXTRACT(YEAR FROM created_at) = ?", year)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("summary ILIKE ?", searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting
	sortBy := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		switch sort {
		case "rating":
			sortBy = "view_count DESC" // Would need rating field
		case "views":
			sortBy = "view_count DESC"
		case "date":
			sortBy = "created_at DESC"
		}
	}
	query = query.Order(sortBy)

	// Apply pagination
	if page, ok := filters["page"].(int); ok {
		if limit, ok := filters["limit"].(int); ok {
			offset := (page - 1) * limit
			query = query.Offset(offset).Limit(limit)
		}
	}

	// Preload relationships
	err := query.
		Preload("Team.Members.User").
		Preload("Proposal.Advisor").
		Preload("Department").
		Preload("Proposal.Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC")
		}).
		Find(&projects).Error

	return projects, int(total), err
}

func (r *repository) GetByAdvisor(advisorID uint) ([]domain.Project, error) {
	var projects []domain.Project
	err := r.db.
		Preload("Team.Members.User").
		Preload("Proposal.Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC")
		}).
		Joins("JOIN proposals ON proposals.id = projects.proposal_id").
		Where("proposals.advisor_id = ?", advisorID).
		Find(&projects).Error
	return projects, err
}

