package domain

import (
	"time"

	"backend/pkg/enums"
)

type University struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Name             string     `gorm:"unique;not null" json:"name"`
	AcademicYear     string     `gorm:"type:varchar(50)" json:"academic_year"`
	ProjectPeriod    string     `gorm:"type:varchar(100)" json:"project_period"`
	VisibilityRule   string     `gorm:"type:varchar(50);default:'private'" json:"visibility_rule"` // private, public, restricted
	AICheckerEnabled bool       `gorm:"default:true" json:"ai_checker_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `gorm:"index" json:"-"`
}

type Department struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	Code         string     `gorm:"type:varchar(20)" json:"code"` // e.g., CSE, SE
	UniversityID uint       `json:"university_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
	University   University `gorm:"foreignKey:UniversityID"`
}

type User struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"not null" json:"name"`
	Email               string     `gorm:"unique;not null" json:"email"`
	Password            string     `gorm:"not null" json:"-"`
	Role                enums.Role `gorm:"type:varchar(20);not null" json:"role"`
	UniversityID        uint       `json:"university_id"`
	DepartmentID        uint       `json:"department_id"`
	StudentID           string     `json:"student_id"`
	ProfilePhoto        string     `json:"profile_photo"`
	IsActive            bool       `gorm:"default:true" json:"is_active"`
	EmailVerified       bool       `gorm:"default:false" json:"email_verified"`
	FailedLoginAttempts int        `gorm:"default:0" json:"-"`
	AccountLockedUntil  *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `gorm:"index" json:"-"`
	University          University `gorm:"foreignKey:UniversityID"`
	Department          Department `gorm:"foreignKey:DepartmentID"`
}

type Team struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	Name         string           `gorm:"not null" json:"name"`
	DepartmentID uint             `json:"department_id"`
	CreatedBy    uint             `json:"created_by"`
	AdvisorID    *uint            `json:"advisor_id"`
	Status       enums.TeamStatus `gorm:"type:varchar(30);default:'pending_advisor_approval'" json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	Department   Department       `gorm:"foreignKey:DepartmentID"`
	Creator      User             `gorm:"foreignKey:CreatedBy"`
	Advisor      *User            `gorm:"foreignKey:AdvisorID"`
	Members      []User           `gorm:"many2many:team_members;" json:"members"`
}

type TeamMember struct {
	TeamID           uint                   `gorm:"primaryKey"`
	UserID           uint                   `gorm:"primaryKey"`
	Role             string                 `gorm:"type:varchar(20);not null"` // leader, member
	InvitationStatus enums.InvitationStatus `gorm:"type:varchar(20);default:'pending'"`
}

type Proposal struct {
	ID               uint                 `gorm:"primaryKey" json:"id"`
	TeamID           uint                 `gorm:"uniqueIndex" json:"team_id"`
	Status           enums.ProposalStatus `gorm:"type:varchar(30);default:'draft'" json:"status"`
	CurrentVersionID *uint                `json:"current_version_id"`
	SubmittedAt      *time.Time           `json:"submitted_at"`
	SubmissionCount  int                  `gorm:"default:0" json:"submission_count"`
	ApprovedAt       *time.Time           `json:"approved_at"`
	ApprovedBy       *uint                `json:"approved_by"`
	RejectedAt       *time.Time           `json:"rejected_at"`
	RejectedBy       *uint                `json:"rejected_by"`
	RejectionReason  string               `json:"rejection_reason"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	DeletedAt        *time.Time           `gorm:"index" json:"-"`
	Team             Team                 `gorm:"foreignKey:TeamID"`
	CurrentVersion   *ProposalVersion     `gorm:"foreignKey:CurrentVersionID"`
	Versions         []ProposalVersion    `json:"versions"`
	Feedback         []Feedback           `json:"feedback"`
	Approver         *User                `gorm:"foreignKey:ApprovedBy"`
	Rejecter         *User                `gorm:"foreignKey:RejectedBy"`
}

type ProposalVersion struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProposalID        uint      `gorm:"index" json:"proposal_id"`
	Title             string    `gorm:"type:varchar(500);not null" json:"title"`
	Objectives        string    `gorm:"type:text;not null" json:"objectives"`
	Methodology       string    `gorm:"type:text;not null" json:"methodology"`
	ExpectedOutcomes  string    `gorm:"type:text;not null" json:"expected_outcomes"`
	FileURL           string    `gorm:"type:varchar(500);not null" json:"file_url"`
	FileHash          string    `gorm:"type:varchar(64);not null" json:"file_hash"`
	FileSizeBytes     int64     `gorm:"not null" json:"file_size_bytes"`
	VersionNumber     int       `gorm:"not null" json:"version_number"`
	IsApprovedVersion bool      `gorm:"default:false" json:"is_approved_version"`
	CreatedBy         uint      `gorm:"not null" json:"created_by"`
	IPAddress         string    `gorm:"type:inet" json:"-"`
	UserAgent         string    `gorm:"type:text" json:"-"`
	SessionID         string    `gorm:"type:varchar(255)" json:"-"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Proposal          Proposal  `gorm:"foreignKey:ProposalID"`
	Creator           User      `gorm:"foreignKey:CreatedBy"`
}

type Feedback struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	ProposalID        uint             `gorm:"index" json:"proposal_id"`
	ProposalVersionID uint             `gorm:"index" json:"proposal_version_id"`
	ReviewerID        uint             `gorm:"index" json:"reviewer_id"`
	Decision          FeedbackDecision `gorm:"type:varchar(20);not null" json:"decision"`
	Comment           string           `gorm:"type:text;not null" json:"comment"`
	IsStructured      bool             `gorm:"default:false" json:"is_structured"`
	IPAddress         string           `gorm:"type:inet" json:"-"`
	UserAgent         string           `gorm:"type:text" json:"-"`
	SessionID         string           `gorm:"type:varchar(255)" json:"-"`
	CreatedAt         time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Proposal          Proposal         `gorm:"foreignKey:ProposalID"`
	Version           ProposalVersion  `gorm:"foreignKey:ProposalVersionID"`
	Reviewer          User             `gorm:"foreignKey:ReviewerID"`
}

type FeedbackDecision string

const (
	FeedbackDecisionApprove FeedbackDecision = "approve"
	FeedbackDecisionRevise  FeedbackDecision = "revise"
	FeedbackDecisionReject  FeedbackDecision = "reject"
)

type Project struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProposalID   uint       `gorm:"unique" json:"proposal_id"`
	TeamID       uint       `json:"team_id"`
	Summary      string     `json:"summary"`
	ApprovedBy   uint       `json:"approved_by"`
	DepartmentID uint       `json:"department_id"`
	Visibility   string     `gorm:"type:varchar(20);default:'private'" json:"visibility"`
	ShareCount   int        `gorm:"default:0" json:"share_count"`
	CreatedAt    time.Time  `json:"created_at"`
	Proposal     Proposal   `gorm:"foreignKey:ProposalID"`
	Team         Team       `gorm:"foreignKey:TeamID"`
	Department   Department `gorm:"foreignKey:DepartmentID"`
	Approver     *User      `gorm:"foreignKey:ApprovedBy"`
}

type ProjectDocumentation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `json:"project_id"`
	DocumentType  string    `gorm:"type:varchar(30)" json:"document_type"` // final_report, etc.
	FileURL       string    `json:"file_url"`
	Status        string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReviewComment string    `json:"review_comment"`
	ReviewedBy    uint      `json:"reviewed_by"`
	ReviewedAt    time.Time `json:"reviewed_at"`
	SubmittedBy   uint      `json:"submitted_by"`
	SubmittedAt   time.Time `json:"submitted_at"`
}

type ProjectReview struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `json:"project_id"`
	UserID    uint      `json:"user_id"`
	Rate      int       `json:"rate"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type Notification struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index" json:"user_id"`
	ReferenceType string     `gorm:"type:varchar(50);not null" json:"reference_type"`
	ReferenceID   uint       `json:"reference_id"`
	Title         string     `gorm:"type:varchar(255);not null" json:"title"`
	Message       string     `gorm:"type:text;not null" json:"message"`
	ActionURL     string     `gorm:"type:varchar(500)" json:"action_url"`
	IsRead        bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt        *time.Time `json:"read_at"`
	Priority      string     `gorm:"type:varchar(20);default:'normal'" json:"priority"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User          User       `gorm:"foreignKey:UserID"`
}

// AuditLog represents system-wide audit trail (immutable)
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EntityType string    `gorm:"type:varchar(50);not null;index" json:"entity_type"`
	EntityID   uint      `gorm:"index" json:"entity_id"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	ActorID    *uint     `gorm:"index" json:"actor_id"`
	ActorRole  string    `gorm:"type:varchar(20)" json:"actor_role"`
	ActorEmail string    `gorm:"type:varchar(255)" json:"actor_email"`
	OldState   string    `gorm:"type:jsonb" json:"old_state"`
	NewState   string    `gorm:"type:jsonb" json:"new_state"`
	Changes    string    `gorm:"type:jsonb" json:"changes"`
	IPAddress  string    `gorm:"type:inet" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	RequestID  string    `gorm:"type:varchar(255)" json:"request_id"`
	SessionID  string    `gorm:"type:varchar(255);index" json:"session_id"`
	Timestamp  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"timestamp"`
	Metadata   string    `gorm:"type:jsonb" json:"metadata"`
	Actor      *User     `gorm:"foreignKey:ActorID"`
}
