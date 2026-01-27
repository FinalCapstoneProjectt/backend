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

// 1. Create Team
func (s *Service) CreateTeam(name string, creatorID uint, deptID uint) (*domain.Team, error) {
	team := &domain.Team{
		Name:         name,
		DepartmentID: deptID,
		CreatedBy:    creatorID,
		IsFinalized:  false,
		AdvisorID:    nil,
	}

	if err := s.repo.CreateWithLeader(team, creatorID); err != nil {
		return nil, err
	}

	// 👇 NEW: Fetch the creator's details to populate the response
	var creator domain.User
	// We access the DB directly here for speed, or you can add GetUser to Repo
	if err := s.repo.GetDB().First(&creator, creatorID).Error; err == nil {
		// Clear sensitive data
		creator.Password = "" 
		
		// Manually attach the full user object
		team.Members = []domain.TeamMember{
			{
				TeamID:           team.ID,
				UserID:           creatorID,
				Role:             "leader",
				InvitationStatus: enums.InvitationStatusAccepted,
				User:             creator, // <--- THIS FILLS THE DATA
			},
		}
	}

	return team, nil
}
// 2. Invite Member
func (s *Service) InviteMember(teamID, inviteeID, requesterID uint) error {
	// A. Check Team Existence
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return err
	}

	// B. Rule: Team Locked?
	if team.IsFinalized {
		return errors.New("cannot invite members: team is finalized")
	}

	// C. Rule: Only Leader can invite
	if !s.isLeader(team, requesterID) {
		return errors.New("only team leader can invite members")
	}

	// D. Add to DB
	member := &domain.TeamMember{
		TeamID:           teamID,
		UserID:           inviteeID,
		Role:             "member",
		InvitationStatus: enums.InvitationStatusPending,
	}
	return s.repo.AddMember(member)
}

// 3. Respond to Invite
func (s *Service) RespondToInvitation(teamID, userID uint, accept bool) error {
	if !accept {
		return s.repo.RemoveMember(teamID, userID)
	}
	return s.repo.UpdateMemberStatus(teamID, userID, enums.InvitationStatusAccepted)
}

// 4. Finalize Team (The Lock)
func (s *Service) FinalizeTeam(teamID, requesterID uint) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return err
	}

	if !s.isLeader(team, requesterID) {
		return errors.New("only team leader can finalize the team")
	}
	
	// Optional: Check min members count here
	if len(team.Members) < 1 {
		return errors.New("team must have members to finalize")
	}

	team.IsFinalized = true
	return s.repo.Update(team)
}

// Helper
func (s *Service) isLeader(team *domain.Team, userID uint) bool {
	for _, m := range team.Members {
		if m.UserID == userID && m.Role == "leader" {
			return true
		}
	}
	return false
}

// Getters for Handler
func (s *Service) GetMyTeams(userID uint, availableOnly bool) ([]domain.Team, error) {
	return s.repo.GetByUserID(userID, availableOnly)
}

func (s *Service) GetTeam(id uint) (*domain.Team, error) {
	return s.repo.GetByID(id)
}

// GetTeamMembers retrieves the list of users in a team
func (s *Service) GetTeamMembers(teamID uint) ([]domain.User, error) {
	// 1. Get the team (Repo already preloads Members and Members.User)
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	// 2. Extract the User objects from the TeamMember relationship
	var users []domain.User
	for _, member := range team.Members {
		// Verify user data exists (safety check)
		if member.User.ID != 0 {
			users = append(users, member.User)
		}
	}

	return users, nil
}

func (s *Service) RemoveMember(teamID, memberID, requesterID uint) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil { return err }

	// Rule: Cannot remove if finalized
	if team.IsFinalized {
		return errors.New("cannot remove members: team is finalized")
	}

	// Rule: Only leader can remove others
	if !s.isLeader(team, requesterID) {
		return errors.New("only team leader can remove members")
	}

	// Rule: Leader cannot remove themselves via this method (must delete team or transfer)
	if memberID == requesterID {
		return errors.New("leader cannot remove themselves, delete team instead")
	}

	return s.repo.RemoveMember(teamID, memberID)
}

// 6. Transfer Leadership
func (s *Service) TransferLeadership(teamID, currentLeaderID, newLeaderID uint) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil { return err }

	// Rule: Cannot transfer if finalized (Strict rule, or optional based on your pref)
	if team.IsFinalized {
		return errors.New("cannot transfer leadership: team is finalized")
	}

	// Verify Requester is Leader
	if !s.isLeader(team, currentLeaderID) {
		return errors.New("unauthorized action")
	}

	// Verify New Leader is in the team
	isMember := false
	for _, m := range team.Members {
		if m.UserID == newLeaderID && m.InvitationStatus == enums.InvitationStatusAccepted {
			isMember = true
			break
		}
	}
	if !isMember {
		return errors.New("new leader must be an active member of the team")
	}

	// Perform Swap (Ideally in Transaction, but doing step-by-step for simplicity)
	// 1. Demote Old Leader
	if err := s.repo.UpdateMemberRole(teamID, currentLeaderID, "member"); err != nil {
		return err
	}
	// 2. Promote New Leader
	if err := s.repo.UpdateMemberRole(teamID, newLeaderID, "leader"); err != nil {
		// Rollback logic would go here in production
		return err
	}
	
	// Update Team CreatedBy field? Optional, but role is more important.
	return nil
}

// 7. Delete Team
func (s *Service) DeleteTeam(teamID, requesterID uint) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil { return err }

	// Rule: Only Leader
	if !s.isLeader(team, requesterID) {
		return errors.New("only team leader can delete the team")
	}

	// Rule: Cannot delete if finalized
	if team.IsFinalized {
		return errors.New("cannot delete a finalized team")
	}

	// Rule: Cannot delete if Proposal exists
	if len(team.Proposals) > 0 {
		return errors.New("cannot delete team: a proposal has already been created")
	}

	return s.repo.Delete(teamID)
}

// 8. Assign Advisor
func (s *Service) AssignAdvisor(teamID, requesterID, advisorID uint) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return err
	}

	// Rule: Only Leader can assign
	if !s.isLeader(team, requesterID) {
		return errors.New("only team leader can assign advisor")
	}

	// Rule: Cannot change advisor if finalized
	if team.IsFinalized {
		return errors.New("cannot change advisor: team is finalized")
	}

	return s.repo.AssignAdvisor(teamID, advisorID)
}

// 9. Advisor Response (approve/reject team assignment)
func (s *Service) AdvisorResponse(teamID, advisorID uint, decision, comment string) error {
	team, err := s.repo.GetByID(teamID)
	if err != nil {
		return err
	}

	// Rule: Only assigned advisor can respond
	if team.AdvisorID == nil || *team.AdvisorID != advisorID {
		return errors.New("only assigned advisor can respond")
	}

	// Apply decision
	if decision == "approve" {
		// Approve the team - can now create proposals
		team.IsFinalized = true
		return s.repo.Update(team)
	} else {
		// Reject - remove advisor assignment
		return s.repo.RemoveAdvisor(teamID)
	}
}