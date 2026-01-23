package proposals

import (
	"backend/pkg/enums"
)

// CanEdit checks if the proposal content can be changed
func CanEdit(status enums.ProposalStatus) bool {
	switch status {
	case enums.ProposalStatusDraft, 
	     enums.ProposalStatusRevisionRequired, 
	     enums.ProposalStatusRejected:
		return true
	default:
		// Submitted, UnderReview, Approved -> LOCKED
		return false
	}
}

// CanSubmit checks if the proposal can be submitted to an advisor
func CanSubmit(status enums.ProposalStatus) bool {
	switch status {
	case enums.ProposalStatusDraft, 
	     enums.ProposalStatusRevisionRequired, 
	     enums.ProposalStatusRejected:
		return true
	default:
		return false
	}
}