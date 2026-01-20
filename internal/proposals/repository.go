package proposals

import (
	"backend/internal/domain"
	"backend/pkg/enums"

	"gorm.io/gorm"
)

type Repository interface {
	Create(proposal *domain.Proposal) error
	GetByID(id uint) (*domain.Proposal, error)
	GetByTeamID(teamID uint) (*domain.Proposal, error)
	GetAll(filters map[string]interface{}) ([]domain.Proposal, error)
	Update(proposal *domain.Proposal) error
	Delete(id uint) error
	CreateVersion(version *domain.ProposalVersion) error
	GetVersionsByProposalID(proposalID uint) ([]domain.ProposalVersion, error)
	GetVersionByID(versionID uint) (*domain.ProposalVersion, error)
	UpdateProposalStatus(id uint, status enums.ProposalStatus) error
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
	err := r.db.Preload("Team").Preload("Team.Members").Preload("Team.Department").Preload("Versions").Preload("CurrentVersion").Preload("Feedback").Preload("Feedback.Reviewer").First(&proposal, id).Error
	if err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (r *repository) GetByTeamID(teamID uint) (*domain.Proposal, error) {
	var proposal domain.Proposal
	err := r.db.Preload("Team").Preload("Versions").Preload("CurrentVersion").Where("team_id = ?", teamID).First(&proposal).Error
	if err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (r *repository) GetAll(filters map[string]interface{}) ([]domain.Proposal, error) {
	var proposals []domain.Proposal
	query := r.db.Preload("Team").Preload("Team.Department").Preload("CurrentVersion")

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if teamID, ok := filters["team_id"]; ok {
		query = query.Where("team_id = ?", teamID)
	}
	if departmentID, ok := filters["department_id"]; ok {
		query = query.Joins("JOIN teams ON proposals.team_id = teams.id").Where("teams.department_id = ?", departmentID)
	}

	err := query.Find(&proposals).Error
	return proposals, err
}

func (r *repository) Update(proposal *domain.Proposal) error {
	return r.db.Save(proposal).Error
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

func (r *repository) GetVersionByID(versionID uint) (*domain.ProposalVersion, error) {
	var version domain.ProposalVersion
	err := r.db.First(&version, versionID).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *repository) UpdateProposalStatus(id uint, status enums.ProposalStatus) error {
	return r.db.Model(&domain.Proposal{}).Where("id = ?", id).Update("status", status).Error
}
