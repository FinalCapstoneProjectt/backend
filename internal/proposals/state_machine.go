package proposals

import (
	"backend/pkg/enums"
	"errors"
	"fmt"
)

// StateTransition defines allowed state changes
type StateTransition struct {
	From []enums.ProposalStatus
	To   enums.ProposalStatus
}

// ValidTransitions maps all allowed state transitions
var ValidTransitions = map[enums.ProposalStatus][]enums.ProposalStatus{
	enums.ProposalStatusDraft: {
		enums.ProposalStatusSubmitted,
	},
	enums.ProposalStatusSubmitted: {
		enums.ProposalStatusUnderReview,
	},
	enums.ProposalStatusUnderReview: {
		enums.ProposalStatusRevisionRequired,
		enums.ProposalStatusApproved,
		enums.ProposalStatusRejected,
	},
	enums.ProposalStatusRevisionRequired: {
		enums.ProposalStatusDraft, // When new version created
	},
	// Terminal states have no transitions
	enums.ProposalStatusApproved: {},
	enums.ProposalStatusRejected: {},
}

// CanTransition checks if state transition is valid
func CanTransition(from, to enums.ProposalStatus) bool {
	allowedStates, exists := ValidTransitions[from]
	if !exists {
		return false
	}

	for _, state := range allowedStates {
		if state == to {
			return true
		}
	}
	return false
}

// ValidateTransition validates and returns error if invalid
func ValidateTransition(from, to enums.ProposalStatus) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid state transition from %s to %s", from, to)
	}
	return nil
}

// IsEditable checks if proposal can be edited
func IsEditable(status enums.ProposalStatus) bool {
	return status == enums.ProposalStatusDraft
}

// CanCreateVersion checks if new version can be created
func CanCreateVersion(status enums.ProposalStatus) bool {
	return status == enums.ProposalStatusDraft ||
		status == enums.ProposalStatusRevisionRequired
}

// CanSubmit checks if proposal can be submitted
func CanSubmit(status enums.ProposalStatus) bool {
	return status == enums.ProposalStatusDraft
}

// CanReview checks if proposal can be reviewed
func CanReview(status enums.ProposalStatus) bool {
	return status == enums.ProposalStatusSubmitted ||
		status == enums.ProposalStatusUnderReview
}

// IsTerminal checks if state is terminal (no further transitions)
func IsTerminal(status enums.ProposalStatus) bool {
	return status == enums.ProposalStatusApproved ||
		status == enums.ProposalStatusRejected
}

// GetAllowedTransitions returns all possible next states
func GetAllowedTransitions(current enums.ProposalStatus) []enums.ProposalStatus {
	return ValidTransitions[current]
}

// GetStateDescription returns human-readable description of state
func GetStateDescription(status enums.ProposalStatus) string {
	descriptions := map[enums.ProposalStatus]string{
		enums.ProposalStatusDraft:            "Draft - Can be edited and submitted",
		enums.ProposalStatusSubmitted:        "Submitted - Awaiting advisor review",
		enums.ProposalStatusUnderReview:      "Under Review - Being evaluated by advisor",
		enums.ProposalStatusRevisionRequired: "Revision Required - Must create new version",
		enums.ProposalStatusApproved:         "Approved - Project created",
		enums.ProposalStatusRejected:         "Rejected - Terminal state",
	}
	return descriptions[status]
}

// StatePermissions defines what actions are allowed in each state
type StatePermissions struct {
	CanEdit          bool
	CanSubmit        bool
	CanCreateVersion bool
	CanReview        bool
	CanApprove       bool
	CanReject        bool
}

// GetStatePermissions returns permissions for a given state
func GetStatePermissions(status enums.ProposalStatus) StatePermissions {
	switch status {
	case enums.ProposalStatusDraft:
		return StatePermissions{
			CanEdit:          true,
			CanSubmit:        true,
			CanCreateVersion: true,
			CanReview:        false,
			CanApprove:       false,
			CanReject:        false,
		}
	case enums.ProposalStatusSubmitted:
		return StatePermissions{
			CanEdit:          false,
			CanSubmit:        false,
			CanCreateVersion: false,
			CanReview:        true,
			CanApprove:       false,
			CanReject:        false,
		}
	case enums.ProposalStatusUnderReview:
		return StatePermissions{
			CanEdit:          false,
			CanSubmit:        false,
			CanCreateVersion: false,
			CanReview:        true,
			CanApprove:       true,
			CanReject:        true,
		}
	case enums.ProposalStatusRevisionRequired:
		return StatePermissions{
			CanEdit:          false,
			CanSubmit:        false,
			CanCreateVersion: true,
			CanReview:        false,
			CanApprove:       false,
			CanReject:        false,
		}
	default:
		// Terminal states or unknown
		return StatePermissions{}
	}
}

// ValidateAction checks if an action is allowed in current state
func ValidateAction(status enums.ProposalStatus, action string) error {
	permissions := GetStatePermissions(status)

	switch action {
	case "edit":
		if !permissions.CanEdit {
			return errors.New("cannot edit proposal in current state")
		}
	case "submit":
		if !permissions.CanSubmit {
			return errors.New("cannot submit proposal in current state")
		}
	case "create_version":
		if !permissions.CanCreateVersion {
			return errors.New("cannot create new version in current state")
		}
	case "review":
		if !permissions.CanReview {
			return errors.New("cannot review proposal in current state")
		}
	case "approve":
		if !permissions.CanApprove {
			return errors.New("cannot approve proposal in current state")
		}
	case "reject":
		if !permissions.CanReject {
			return errors.New("cannot reject proposal in current state")
		}
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}
