package users

import (
	"backend/internal/domain"
	"backend/pkg/enums"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

type CreateTeacherRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	UniversityID uint   `json:"university_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
}

type CreateStudentRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	StudentID    string `json:"student_id" binding:"required"`
	UniversityID uint   `json:"university_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type AssignDepartmentRequest struct {
	DepartmentID uint `json:"department_id" binding:"required"`
}

type UserResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         enums.Role `json:"role"`
	StudentID    string     `json:"student_id,omitempty"`
	UniversityID uint       `json:"university_id"`
	DepartmentID uint       `json:"department_id"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    string     `json:"created_at"`
}

func (s *Service) CreateTeacher(req CreateTeacherRequest) (*domain.User, error) {
	// Check if email already exists
	existing, _ := s.repo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Role:         enums.RoleAdvisor,
		UniversityID: req.UniversityID,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) CreateStudent(req CreateStudentRequest) (*domain.User, error) {
	// Check if email already exists
	existing, _ := s.repo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Role:         enums.RoleStudent,
		StudentID:    req.StudentID,
		UniversityID: req.UniversityID,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) GetUser(id uint) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAllUsers(role string, departmentID uint, universityID uint, isActive *bool) ([]domain.User, error) {
	filters := make(map[string]interface{})

	if role != "" {
		filters["role"] = role
	}
	if departmentID > 0 {
		filters["department_id"] = departmentID
	}
	if universityID > 0 {
		filters["university_id"] = universityID
	}
	if isActive != nil {
		filters["is_active"] = *isActive
	}

	return s.repo.GetAll(filters)
}

func (s *Service) UpdateUserStatus(id uint, isActive bool) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.UpdateStatus(id, isActive)
}

func (s *Service) AssignDepartment(userID uint, departmentID uint) error {
	_, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.AssignDepartment(userID, departmentID)
}

func (s *Service) DeleteUser(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	return s.repo.Delete(id)
}

// Add Implementation
func (s *Service) GetPeers(departmentID uint, universityID uint, excludeUserID uint) ([]domain.User, error) {
	return s.repo.FindPeers(departmentID, universityID, excludeUserID)
}

// Add DTO
type AdvisorWorkload struct {
    Advisor   domain.User `json:"advisor"`
	Proposals []domain.Proposal `json:"proposals"` 
    TeamCount int64       `json:"team_count"`
}

// Add Method to Service Interface/Struct
func (s *Service) GetDepartmentAdvisorsWithWorkload(departmentID uint) ([]AdvisorWorkload, error) {
    advisors, err := s.repo.GetAdvisorsByDepartment(departmentID)
    if err != nil {
        return nil, err
    }

    var result []AdvisorWorkload
    for _, adv := range advisors {
        var assignedProposals []domain.Proposal
        
        // Fetch proposals for THIS advisor, preloading Team and Latest Version
        s.repo.GetDB().
            Preload("Team").
            Preload("Versions", "version_number = 1").
            Where("advisor_id = ?", adv.ID).
            Find(&assignedProposals)

        adv.Password = "" // Security
        result = append(result, AdvisorWorkload{
            Advisor:   adv,
            TeamCount: int64(len(assignedProposals)),
            Proposals: assignedProposals,
        })
    }
    
    return result, nil
}

type AdminDashboardStats struct {
    PendingCount      int64             `json:"pending_assignment"`
    UnderReviewCount  int64             `json:"under_review"`
    ApprovedCount     int64             `json:"approved"`
    TotalTeams        int64             `json:"total_teams"`
    AvailableAdvisors int64             `json:"available_advisors"`
    RecentProposals   []domain.Proposal `json:"recent_proposals"`
    AdvisorWorkload   []AdvisorWorkload `json:"advisor_workload"`
}

// Service Method
func (s *Service) GetAdminDashboardStats(deptID uint) (*AdminDashboardStats, error) {
    stats := &AdminDashboardStats{}

	    // FIX 1: Approved Count Query
    s.repo.GetDB().Model(&domain.Proposal{}).
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ? AND proposals.status = ?", deptID, enums.ProposalStatusApproved).
        Count(&stats.ApprovedCount)

    // FIX 2: Preload Leader in RecentProposals
    s.repo.GetDB().
        Preload("Team").
        Preload("Team.Members.User"). // 👈 FIX: Load Users inside Members
        Preload("Versions", "version_number = 1").
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ?", deptID). // 👈 FIX: Show ALL department proposals, not just submitted
        Order("proposals.created_at DESC").
        Limit(10). // Increased limit
        Find(&stats.RecentProposals)
    
    // 1. Proposal Counts (Using raw SQL or multiple count queries for speed)
    s.repo.GetDB().Model(&domain.Proposal{}).
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ? AND proposals.status = ?", deptID, enums.ProposalStatusSubmitted).
        Count(&stats.PendingCount)

    s.repo.GetDB().Model(&domain.Proposal{}).
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ? AND proposals.status = ?", deptID, enums.ProposalStatusUnderReview).
        Count(&stats.UnderReviewCount)

    s.repo.GetDB().Model(&domain.Proposal{}).
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ? AND proposals.status = ?", deptID, enums.ProposalStatusApproved).
        Count(&stats.ApprovedCount)

    s.repo.GetDB().Model(&domain.Team{}).
        Where("department_id = ?", deptID).
        Count(&stats.TotalTeams)

    // 2. Recent Pending Proposals (Limit 5)
    s.repo.GetDB().
        Preload("Team").
        Preload("Versions", "version_number = 1"). // Get Title
        Joins("JOIN teams ON teams.id = proposals.team_id").
        Where("teams.department_id = ? AND proposals.status = ?", deptID, enums.ProposalStatusSubmitted).
        Order("proposals.created_at DESC").
        Limit(5).
        Find(&stats.RecentProposals)

    // 3. Advisor Workload (Reuse existing logic)
    workload, _ := s.GetDepartmentAdvisorsWithWorkload(deptID)
    stats.AdvisorWorkload = workload
    
    // Calc Available Advisors (Capacity > Workload)
    // Assuming hardcoded capacity of 5 for now
    for _, w := range workload {
        if w.TeamCount < 5 {
            stats.AvailableAdvisors++
        }
    }

    return stats, nil
}