package proposals

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Repository interface {
	Create(proposal *domain.Proposal) error
	GetByID(id uint) (*domain.Proposal, error)
	GetAll(filters map[string]interface{}) ([]domain.Proposal, error)
	Update(proposal *domain.Proposal) error
	Delete(id uint) error
	
	// Versioning
	CreateVersion(version *domain.ProposalVersion) error
	GetVersionsByProposalID(proposalID uint) ([]domain.ProposalVersion, error)
	GetLatestVersion(proposalID uint) (*domain.ProposalVersion, error)
	GetFirstVersion(proposalID uint) (*domain.ProposalVersion, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(proposal *domain.Proposal) error {
	return r.db.Create(proposal).Error
}

func (r *repository) GetByID(id uint) (*domain.Proposal, error) {
	var proposal domain.Proposal
	
	// 👇 CRITICAL: Preload Versions with Order
	err := r.db.
		Preload("Team").
		Preload("Team.Members.User"). // Load team members for display
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC") // Latest first!
		}).
		First(&proposal, id).Error
		
	if err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (r *repository) GetAll(filters map[string]interface{}) ([]domain.Proposal, error) {
	var proposals []domain.Proposal
	query := r.db.Preload("Team").
        Preload("Team.Department").
        Preload("Team.Members"). // To count team size
        Preload("Versions", func(db *gorm.DB) *gorm.DB {
            return db.Order("version_number DESC") // Get latest version first
        })

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if departmentID, ok := filters["department_id"]; ok {
		query = query.Joins("JOIN teams ON proposals.team_id = teams.id").
			Where("teams.department_id = ?", departmentID)
	}

	err := query.Find(&proposals).Error
	return proposals, err
}

func (r *repository) Update(proposal *domain.Proposal) error {
	return r.db.Omit("Team", "Versions", "CurrentVersion", "Feedback").Save(proposal).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Proposal{}, id).Error
}

func (r *repository) CreateVersion(version *domain.ProposalVersion) error {
	return r.db.Create(version).Error
}

func (r *repository) GetVersionsByProposalID(proposalID uint) ([]domain.ProposalVersion, error) {
	var versions []domain.ProposalVersion
	err := r.db.Where("proposal_id = ?", proposalID).Order("version_number DESC").Find(&versions).Error
	return versions, err
}

func (r *repository) GetLatestVersion(proposalID uint) (*domain.ProposalVersion, error) {
	var version domain.ProposalVersion
	err := r.db.Where("proposal_id = ?", proposalID).Order("version_number DESC").First(&version).Error
	return &version, err
}

func (r *repository) GetFirstVersion(proposalID uint) (*domain.ProposalVersion, error) {
	var version domain.ProposalVersion
	err := r.db.Where("proposal_id = ? AND version_number = 1", proposalID).First(&version).Error
	return &version, err
}