package teams

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"gorm.io/gorm"
)

type Repository interface {
	CreateWithLeader(team *domain.Team, leaderID uint) error
	GetByID(id uint) (*domain.Team, error)
	GetByUserID(userID uint, availableOnly bool) ([]domain.Team, error)
	Update(team *domain.Team) error
	GetDB() *gorm.DB

	// Member management
	AddMember(member *domain.TeamMember) error
	RemoveMember(teamID, userID uint) error
	GetMember(teamID, userID uint) (*domain.TeamMember, error)
	UpdateMemberStatus(teamID, userID uint, status enums.InvitationStatus) error
	Delete(id uint) error
	UpdateMemberRole(teamID, userID uint, role string) error // <--- Added
}

type repository struct {
	db *gorm.DB
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}


func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CreateWithLeader handles the transaction: Create Team AND Add Leader
func (r *repository) CreateWithLeader(team *domain.Team, leaderID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Team
		if err := tx.Create(team).Error; err != nil {
			return err
		}

		// 2. Add Leader
		leader := domain.TeamMember{
			TeamID:           team.ID,
			UserID:           leaderID,
			Role:             "leader", // You can use an enum here
			InvitationStatus: enums.InvitationStatusAccepted,
		}
		if err := tx.Create(&leader).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repository) GetByID(id uint) (*domain.Team, error) {
	var team domain.Team
	// Added Preload("Proposals") to check for existing proposals before deletion
	err := r.db.Preload("Department").
		Preload("Members.User").
		Preload("Proposals"). 
		First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *repository) Update(team *domain.Team) error {
	return r.db.Save(team).Error
}

func (r *repository) GetByUserID(userID uint, availableOnly bool) ([]domain.Team, error) {
	var teams []domain.Team
	
	query := r.db.Preload("Department").
		Preload("Members").
		Preload("Members.User").
		Preload("Creator").
        Preload("Proposals"). // 👈 Need this to check count
		Joins("JOIN team_members on team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID)

    // Filter Logic
    if availableOnly {
        // Only return teams that have 0 proposals
        // Using GORM subquery or simple client-side filter if list is small.
        // For efficiency, let's use a LEFT JOIN check (teams without proposals)
        query = query.Joins("LEFT JOIN proposals ON proposals.team_id = teams.id").
                Where("proposals.id IS NULL")
    }

	err := query.Find(&teams).Error
	return teams, err
}


func (r *repository) AddMember(member *domain.TeamMember) error {
	return r.db.Create(member).Error
}

func (r *repository) Delete(id uint) error {
	// GORM will handle cascading deletes if setup in DB, 
	// otherwise we delete members first then team.
	// Assuming DB constraints handles cascade or we do soft delete.
	return r.db.Delete(&domain.Team{}, id).Error
}

func (r *repository) RemoveMember(teamID, userID uint) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&domain.TeamMember{}).Error
}

// New: For transferring leadership
func (r *repository) UpdateMemberRole(teamID, userID uint, role string) error {
	return r.db.Model(&domain.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}

func (r *repository) GetMember(teamID, userID uint) (*domain.TeamMember, error) {
	var member domain.TeamMember
	err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *repository) UpdateMemberStatus(teamID, userID uint, status enums.InvitationStatus) error {
	return r.db.Model(&domain.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("invitation_status", status).Error
}