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
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	DepartmentID uint       `json:"department_id"`
	CreatedBy    uint       `json:"created_by"`
	AdvisorID    *uint      `json:"advisor_id"` 
	IsFinalized  bool       `gorm:"default:false" json:"is_finalized"`
	CreatedAt    time.Time  `json:"created_at"`
	
	Department   *Department   `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Creator      *User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Advisor      *User         `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`

	Members      []TeamMember `gorm:"foreignKey:TeamID" json:"members"`
	Proposals    []Proposal   `gorm:"foreignKey:TeamID" json:"proposals"`
}

type TeamMember struct {
	TeamID           uint                   `gorm:"primaryKey" json:"team_id"`
	UserID           uint                   `gorm:"primaryKey" json:"user_id"`
	Role             string                 `gorm:"type:varchar(20);default:'member'" json:"role"` // 'leader', 'member'
	InvitationStatus enums.InvitationStatus `gorm:"type:varchar(20);default:'pending'" json:"invitation_status"`
	
	// Preload User details for UI
	User User `gorm:"foreignKey:UserID" json:"user"`
}

type Proposal struct {
	ID               uint                 `gorm:"primaryKey" json:"id"`
	TeamID           *uint                `json:"team_id"` // ⚠️ Changed to pointer to allow NULL
	AdvisorID        *uint                `json:"advisor_id"`
	Status           enums.ProposalStatus `gorm:"type:varchar(30);default:'draft'" json:"status"`
	CreatedBy         uint   			  `json:"created_by"` // 👈 Add this
	
	// Relationships
	Team             *Team                `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Versions         []ProposalVersion    `gorm:"foreignKey:ProposalID" json:"versions"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	Advisor          *User                `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`

}

// Ensure ProposalVersion matches your DBML
type ProposalVersion struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ProposalID       uint      `json:"proposal_id"`
	Title            string    `json:"title"`
	Abstract         string    `json:"abstract"`
	ProblemStatement string    `json:"problem_statement"`
	Objectives       string    `json:"objectives"`
	Methodology      string    `json:"methodology"`
	ExpectedTimeline string    `json:"expected_timeline"`
	VersionNumber    int       `json:"version_number"`
	ExpectedOutcomes string    `json:"expected_outcomes"`
	FileURL 		*string    `json:"file_url"` //nullable
	IsApproved       bool      `gorm:"default:false" json:"is_approved"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	FileHash      string       `gorm:"type:varchar(64)" json:"file_hash"` // Removed "not null"
    FileSizeBytes int64        `json:"file_size_bytes"`   
	CreatedBy        uint      `json:"created_by"`
    
    // Optional: Relationship
    Creator          User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

type Feedback struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	ProposalID        uint             `gorm:"index" json:"proposal_id"`
	ProposalVersionID uint             `gorm:"index" json:"proposal_version_id"`
	ReviewerID        uint             `gorm:"index" json:"reviewer_id"`
	Decision          FeedbackDecision `gorm:"type:varchar(20);not null" json:"decision"`
	Comment           string           `gorm:"type:text;not null" json:"comment"`
	IsStructured      bool             `gorm:"default:false" json:"is_structured"`
	IPAddress         *string          `gorm:"type:inet" json:"-"`
	UserAgent         *string          `gorm:"type:text" json:"-"`
	SessionID         *string          `gorm:"type:varchar(255)" json:"-"`
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
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProposalID   uint      `gorm:"uniqueIndex" json:"proposal_id"`
	TeamID       uint      `json:"team_id"`
	Summary      string    `json:"summary"`
	ApprovedBy   uint      `json:"approved_by"`
	DepartmentID uint      `json:"department_id"`
	Visibility   string    `gorm:"type:varchar(20);default:'private'" json:"visibility"`
	ShareCount   int       `gorm:"default:0" json:"share_count"`
	CreatedAt    time.Time `json:"created_at"`
	ViewCount    int       `gorm:"default:0" json:"view_count"` // 👈 ADD THIS

	// 👇 ADD THESE RELATIONSHIPS
	Proposal   Proposal   `gorm:"foreignKey:ProposalID" json:"proposal"`
	Team       Team       `gorm:"foreignKey:TeamID" json:"team"`
	Department Department `gorm:"foreignKey:DepartmentID" json:"department"`
	Approver   User       `gorm:"foreignKey:ApprovedBy" json:"approver"`
	
}

type ProjectDocumentation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `json:"project_id"`
	DocumentType  string    `gorm:"type:varchar(30)" json:"document_type"`
	URL           string    `gorm:"column:url" json:"url"` 
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
	Changes    string   `gorm:"type:text" json:"changes"` 
	IPAddress  string    `gorm:"type:inet" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	RequestID  string    `gorm:"type:varchar(255)" json:"request_id"`
	SessionID  string    `gorm:"type:varchar(255);index" json:"session_id"`
	Timestamp  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"timestamp"`
	Metadata   string    `gorm:"type:text" json:"metadata"`
	Actor      *User     `gorm:"foreignKey:ActorID"`
}
