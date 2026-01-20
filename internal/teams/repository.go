package teams

import (
	"backend/internal/domain"
	"backend/pkg/enums"

	"gorm.io/gorm"
)

type Repository interface {
	Create(team *domain.Team) error
	GetByID(id uint) (*domain.Team, error)
	GetByCreator(creatorID uint) ([]domain.Team, error)
	GetByDepartment(departmentID uint) ([]domain.Team, error)
	Update(team *domain.Team) error
	Delete(id uint) error
	AddMember(teamID uint, userID uint, role string, status enums.InvitationStatus) error
	UpdateMemberStatus(teamID uint, userID uint, status enums.InvitationStatus) error
	RemoveMember(teamID uint, userID uint) error
	GetMembers(teamID uint) ([]domain.User, error)
	GetMemberRole(teamID uint, userID uint) (string, error)
	IsMember(teamID uint, userID uint) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(team *domain.Team) error {
	return r.db.Create(team).Error
}

func (r *repository) GetByID(id uint) (*domain.Team, error) {
	var team domain.Team
	err := r.db.Preload("Department").Preload("Creator").Preload("Advisor").Preload("Members").First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *repository) GetByCreator(creatorID uint) ([]domain.Team, error) {
	var teams []domain.Team
	err := r.db.Preload("Department").Preload("Members").Where("created_by = ?", creatorID).Find(&teams).Error
	return teams, err
}

func (r *repository) GetByDepartment(departmentID uint) ([]domain.Team, error) {
	var teams []domain.Team
	err := r.db.Preload("Department").Preload("Creator").Preload("Members").Where("department_id = ?", departmentID).Find(&teams).Error
	return teams, err
}

func (r *repository) Update(team *domain.Team) error {
	return r.db.Save(team).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Team{}, id).Error
}

func (r *repository) AddMember(teamID uint, userID uint, role string, status enums.InvitationStatus) error {
	member := domain.TeamMember{
		TeamID:           teamID,
		UserID:           userID,
		Role:             role,
		InvitationStatus: status,
	}
	return r.db.Create(&member).Error
}

func (r *repository) UpdateMemberStatus(teamID uint, userID uint, status enums.InvitationStatus) error {
	return r.db.Model(&domain.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Update("invitation_status", status).Error
}

func (r *repository) RemoveMember(teamID uint, userID uint) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&domain.TeamMember{}).Error
}

func (r *repository) GetMembers(teamID uint) ([]domain.User, error) {
	var team domain.Team
	err := r.db.Preload("Members").First(&team, teamID).Error
	if err != nil {
		return nil, err
	}
	return team.Members, nil
}

func (r *repository) GetMemberRole(teamID uint, userID uint) (string, error) {
	var member domain.TeamMember
	err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *repository) IsMember(teamID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count).Error
	return count > 0, err
}
