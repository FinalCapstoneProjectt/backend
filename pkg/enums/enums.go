package enums

type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
	RolePublic  Role = "public"
)

type ProposalStatus string

const (
	ProposalStatusDraft            ProposalStatus = "draft"
	ProposalStatusSubmitted        ProposalStatus = "submitted"
	ProposalStatusUnderReview      ProposalStatus = "under_review"
	ProposalStatusRevisionRequired ProposalStatus = "revision_required"
	ProposalStatusApproved         ProposalStatus = "approved"
	ProposalStatusRejected         ProposalStatus = "rejected"
)

type TeamStatus string

const (
	TeamStatusPendingAdvisorApproval TeamStatus = "pending_advisor_approval"
	TeamStatusApproved               TeamStatus = "approved"
	TeamStatusRejected               TeamStatus = "rejected"
)

type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
)
