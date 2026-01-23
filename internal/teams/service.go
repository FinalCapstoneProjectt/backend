package teams

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

type CreateTeamRequest struct {
	Name         string `json:"name" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
	AdvisorID    uint   `json:"advisor_id"`
	MemberIDs    []uint `json:"member_ids"`
}

type InviteMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type RespondInvitationRequest struct {
	Accept bool `json:"accept" binding:"required"`
}

type ApproveTeamRequest struct {
	Approve bool `json:"approve" binding:"required"`
}

func (s *Service) CreateTeam(req CreateTeamRequest, creatorID uint) (*domain.Team, error) {
	team := &domain.Team{
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		CreatedBy:    creatorID,
		Status:       enums.TeamStatusPendingAdvisorApproval,
	}

	// Only set AdvisorID if provided (non-zero)
	if req.AdvisorID != 0 {
		team.AdvisorID = &req.AdvisorID
	}

	err := s.repo.Create(team)
	if err != nil {
		return nil, err
	}

	// Add creator as leader with accepted status
	err = s.repo.AddMember(team.ID, creatorID, "leader", enums.InvitationStatusAccepted)
	if err != nil {
		return nil, err
	}

	// Invite initial members if provided
	for _, memberID := range req.MemberIDs {
		if memberID != creatorID {
			err = s.repo.AddMember(team.ID, memberID, "member", enums.InvitationStatusPending)
			if err != nil {
				// Log error but continue with other members
				continue
			}
		}
	}

	return s.repo.GetByID(team.ID)
}

func (s *Service) GetTeam(id uint) (*domain.Team, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetMyTeams(userID uint) ([]domain.Team, error) {
	return s.repo.GetForUser(userID)
}

func (s *Service) GetTeamMembers(teamID uint) ([]domain.User, error) {
	return s.repo.GetMembers(teamID)
}

func (s *Service) ApproveTeam(teamID uint, advisorID uint, approve bool) (*domain.Team, error) {
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	if team.AdvisorID == nil || *team.AdvisorID != advisorID {
		return nil, errors.New("only assigned advisor can approve this team")
	}

	if team.Status != enums.TeamStatusPendingAdvisorApproval {
		return nil, errors.New("team is not pending advisor approval")
	}

	newStatus := enums.TeamStatusRejected
	if approve {
		newStatus = enums.TeamStatusApproved
	}

	if err := s.repo.UpdateStatus(teamID, newStatus); err != nil {
		return nil, err
	}

	return s.repo.GetByID(teamID)
}

func (s *Service) InviteMember(teamID uint, userID uint, inviterID uint) error {
	// Check if inviter is the leader
	role, err := s.repo.GetMemberRole(teamID, inviterID)
	if err != nil || role != "leader" {
		return errors.New("only team leader can invite members")
	}

	// Check if user is already a member
	isMember, _ := s.repo.IsMember(teamID, userID)
	if isMember {
		return errors.New("user is already a member or has pending invitation")
	}

	return s.repo.AddMember(teamID, userID, "member", enums.InvitationStatusPending)
}

func (s *Service) RespondToInvitation(teamID uint, userID uint, accept bool) error {
	// Check if user has pending invitation
	isMember, err := s.repo.IsMember(teamID, userID)
	if err != nil || !isMember {
		return errors.New("no invitation found")
	}

	if accept {
		return s.repo.UpdateMemberStatus(teamID, userID, enums.InvitationStatusAccepted)
	} else {
		return s.repo.UpdateMemberStatus(teamID, userID, enums.InvitationStatusRejected)
	}
}

func (s *Service) RemoveMember(teamID uint, userID uint, removerID uint) error {
	// Check if remover is the leader
	role, err := s.repo.GetMemberRole(teamID, removerID)
	if err != nil || role != "leader" {
		return errors.New("only team leader can remove members")
	}

	// Prevent leader from removing themselves
	if userID == removerID {
		return errors.New("leader cannot remove themselves")
	}

	return s.repo.RemoveMember(teamID, userID)
}

func (s *Service) IsTeamLeader(teamID uint, userID uint) (bool, error) {
	role, err := s.repo.GetMemberRole(teamID, userID)
	if err != nil {
		return false, err
	}
	return role == "leader", nil
}

func (s *Service) IsTeamMember(teamID uint, userID uint) (bool, error) {
	return s.repo.IsMember(teamID, userID)
}
