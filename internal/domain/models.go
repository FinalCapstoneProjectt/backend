package domain

import (
	"time"

	"backend/pkg/enums"
)

type University struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Department struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	UniversityID uint       `json:"university_id"`
	University   University `gorm:"foreignKey:UniversityID"`
}

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null" json:"name"`
	Email        string     `gorm:"unique;not null" json:"email"`
	Password     string     `gorm:"not null" json:"-"`
	Role         enums.Role `gorm:"type:varchar(20);not null" json:"role"`
	UniversityID uint       `json:"university_id"`
	DepartmentID uint       `json:"department_id"`
	StudentID    string     `json:"student_id"`
	ProfilePhoto string     `json:"profile_photo"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	University   University `gorm:"foreignKey:UniversityID"`
	Department   Department `gorm:"foreignKey:DepartmentID"`
}

type Team struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	Name         string           `gorm:"not null" json:"name"`
	DepartmentID uint             `json:"department_id"`
	CreatedBy    uint             `json:"created_by"`
	AdvisorID    uint             `json:"advisor_id"`
	Status       enums.TeamStatus `gorm:"type:varchar(30);default:'pending_advisor_approval'" json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	Department   Department       `gorm:"foreignKey:DepartmentID"`
	Creator      User             `gorm:"foreignKey:CreatedBy"`
	Advisor      User             `gorm:"foreignKey:AdvisorID"`
	Members      []User           `gorm:"many2many:team_members;" json:"members"`
}

type TeamMember struct {
	TeamID           uint                   `gorm:"primaryKey"`
	UserID           uint                   `gorm:"primaryKey"`
	Role             string                 `gorm:"type:varchar(20);not null"` // leader, member
	InvitationStatus enums.InvitationStatus `gorm:"type:varchar(20);default:'pending'"`
}

type Proposal struct {
	ID        uint                 `gorm:"primaryKey" json:"id"`
	TeamID    uint                 `json:"team_id"`
	Status    enums.ProposalStatus `gorm:"type:varchar(30);default:'draft'" json:"status"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	Team      Team                 `gorm:"foreignKey:TeamID"`
	Versions  []ProposalVersion    `json:"versions"`
}

type ProposalVersion struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProposalID        uint      `json:"proposal_id"`
	Title             string    `gorm:"not null" json:"title"`
	Objectives        string    `json:"objectives"`
	FileURL           string    `json:"file_url"`
	VersionNumber     int       `json:"version_number"`
	IsApprovedVersion bool      `gorm:"default:false" json:"is_approved_version"`
	CreatedAt         time.Time `json:"created_at"`
}

type Feedback struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProposalID        uint      `json:"proposal_id"`
	ProposalVersionID uint      `json:"proposal_version_id"`
	ReviewerID        uint      `json:"reviewer_id"`
	Decision          string    `gorm:"type:varchar(20)" json:"decision"` // approve, revise, reject
	Comment           string    `json:"comment"`
	CreatedAt         time.Time `json:"created_at"`
	Reviewer          User      `gorm:"foreignKey:ReviewerID"`
}

type Project struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProposalID   uint      `gorm:"unique" json:"proposal_id"`
	TeamID       uint      `json:"team_id"`
	Summary      string    `json:"summary"`
	ApprovedBy   uint      `json:"approved_by"`
	DepartmentID uint      `json:"department_id"`
	Visibility   string    `gorm:"type:varchar(20);default:'private'" json:"visibility"`
	ShareCount   int       `gorm:"default:0" json:"share_count"`
	CreatedAt    time.Time `json:"created_at"`
	Proposal     Proposal  `gorm:"foreignKey:ProposalID"`
	Team         Team      `gorm:"foreignKey:TeamID"`
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
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `json:"user_id"`
	ReferenceType string    `json:"reference_type"` // team, proposal, etc.
	ReferenceID   uint      `json:"reference_id"`
	Message       string    `json:"message"`
	IsRead        bool      `gorm:"default:false" json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
}
