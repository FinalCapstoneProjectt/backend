package teams

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}
type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
}

type InviteMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type TransferLeadershipRequest struct {
	NewLeaderID uint `json:"new_leader_id" binding:"required"`
}

type RespondInvitationRequest struct {
	Accept bool `json:"accept"`
}

type AdvisorResponseRequest struct {
	Decision string `json:"decision" binding:"required"` // "approve" or "reject"
	Comment  string `json:"comment" binding:"required,min=10"`
}

type AssignAdvisorRequest struct {
	AdvisorID uint `json:"advisor_id" binding:"required"`
}

// CreateTeam godoc
// @Summary Create a new team
// @Description Student creates a new team and becomes the leader
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param team body CreateTeamRequest true "Team details"
// @Success 201 {object} response.Response{data=domain.Team}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams [post]
func (h *Handler) CreateTeam(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil { return }

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid inputs", err.Error())
		return
	}

	// Pass DepartmentID from Claims!
	team, err := h.service.CreateTeam(req.Name, claims.UserID, claims.DepartmentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create team", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Team created successfully", team)
}

// FinalizeTeam godoc
// @Summary Finalize a team
// @Description Locks the team structure so a proposal can be created. Only Leader can do this.
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /teams/{id}/finalize [post]
func (h *Handler) FinalizeTeam(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil { return }

	teamID := parseID(c)
	if teamID == 0 { return }

	err := h.service.FinalizeTeam(teamID, claims.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to finalize team", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Team finalized successfully", nil)
}

// GetTeams godoc
// @Summary Get user's teams
// @Description Get all teams where the user is a member or creator
// @Tags Teams
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Team}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams [get]
func (h *Handler) GetTeams(c *gin.Context) {
    claims := getClaims(c)
    if claims == nil { return }

    // Check query param
    availableOnly := c.Query("available") == "true"

    teams, err := h.service.GetMyTeams(claims.UserID, availableOnly)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to fetch teams", err.Error())
        return
    }

    response.Success(c, teams)
}

// GetTeam godoc
// @Summary Get team by ID
// @Description Retrieve team details with members
// @Tags Teams
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Success 200 {object} response.Response{data=domain.Team}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /teams/{id} [get]
func (h *Handler) GetTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	team, err := h.service.GetTeam(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Team not found", err.Error())
		return
	}

	response.Success(c, team)
}

// GetTeamMembers godoc
// @Summary Get team members
// @Description Retrieve all members of a team
// @Tags Teams
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Success 200 {object} response.Response{data=[]domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /teams/{id}/members [get]
func (h *Handler) GetTeamMembers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	members, err := h.service.GetTeamMembers(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch team members", err.Error())
		return
	}

	response.Success(c, members)
}

// InviteMember godoc
// @Summary Invite a member to team
// @Description Team leader invites a student to join the team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param invitation body InviteMemberRequest true "User to invite"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams/{id}/invite [post]
func (h *Handler) InviteMember(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.service.InviteMember(uint(id), req.UserID, userClaims.UserID)
	if err != nil {
		if err.Error() == "only team leader can invite members" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to invite member", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Member invited successfully", nil)
}

// RespondToInvitation godoc
// @Summary Respond to team invitation
// @Description Student accepts or rejects a team invitation
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param response body RespondInvitationRequest true "Accept or reject"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams/{id}/invitation/respond [post]
func (h *Handler) RespondToInvitation(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	var req RespondInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.service.RespondToInvitation(uint(id), userClaims.UserID, req.Accept)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to respond to invitation", err.Error())
		return
	}

	message := "Invitation accepted"
	if !req.Accept {
		message = "Invitation rejected"
	}
	response.JSON(c, http.StatusOK, message, nil)
}

// // RemoveMember godoc
// // @Summary Remove a member from team
// // @Description Team leader removes a member from the team
// // @Tags Teams
// // @Produce json
// // @Security BearerAuth
// // @Param id path int true "Team ID"
// // @Param memberId path int true "Member User ID"
// // @Success 200 {object} response.Response
// // @Failure 400 {object} response.ErrorResponse
// // @Failure 401 {object} response.ErrorResponse
// // @Failure 403 {object} response.ErrorResponse
// // @Failure 500 {object} response.ErrorResponse
// // @Router /teams/{id}/members/{memberId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil { return }

	teamID := parseID(c)
	if teamID == 0 { return }

	memberIDString := c.Param("memberId") // Ensure router uses :memberId
	memberID, err := strconv.ParseUint(memberIDString, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid member ID", err.Error())
		return
	}

	err = h.service.RemoveMember(teamID, uint(memberID), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to remove member", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Member removed successfully", nil)
}

// TransferLeadership godoc
// @Summary Transfer team leadership
// @Description Assign a new leader. Old leader becomes a member.
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param request body TransferLeadershipRequest true "New Leader ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /teams/{id}/transfer-leadership [post]
func (h *Handler) TransferLeadership(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil { return }

	teamID := parseID(c)
	if teamID == 0 { return }

	var req TransferLeadershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid inputs", err.Error())
		return
	}

	err := h.service.TransferLeadership(teamID, claims.UserID, req.NewLeaderID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to transfer leadership", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Leadership transferred successfully", nil)
}

// DeleteTeam (New)
func (h *Handler) DeleteTeam(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil { return }

	teamID := parseID(c)
	if teamID == 0 { return }

	err := h.service.DeleteTeam(teamID, claims.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete team", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Team deleted successfully", nil)
}

// AdvisorResponse godoc
// @Summary Advisor responds to team assignment
// @Description Advisor approves or rejects being assigned to a team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param response body AdvisorResponseRequest true "Approval decision"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /teams/{id}/advisor-response [post]
func (h *Handler) AdvisorResponse(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	teamID := parseID(c)
	if teamID == 0 {
		return
	}

	var req AdvisorResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Decision != "approve" && req.Decision != "reject" {
		response.Error(c, http.StatusBadRequest, "Decision must be 'approve' or 'reject'", nil)
		return
	}

	err := h.service.AdvisorResponse(teamID, claims.UserID, req.Decision, req.Comment)
	if err != nil {
		if err.Error() == "only assigned advisor can respond" {
			response.Error(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusBadRequest, "Failed to process advisor response", err.Error())
		return
	}

	message := "Team approved"
	if req.Decision == "reject" {
		message = "Team rejected"
	}
	response.JSON(c, http.StatusOK, message, nil)
}

// AssignAdvisor godoc
// @Summary Assign advisor to team
// @Description Team leader assigns an advisor to the team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param request body AssignAdvisorRequest true "Advisor ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /teams/{id}/assign-advisor [post]
func (h *Handler) AssignAdvisor(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	teamID := parseID(c)
	if teamID == 0 {
		return
	}

	var req AssignAdvisorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.service.AssignAdvisor(teamID, claims.UserID, req.AdvisorID)
	if err != nil {
		if err.Error() == "only team leader can assign advisor" {
			response.Error(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusBadRequest, "Failed to assign advisor", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Advisor assigned successfully", nil)
}

// Helpers
func getClaims(c *gin.Context) *auth.TokenClaims {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return nil
	}
	return claims.(*auth.TokenClaims)
}

func parseID(c *gin.Context) uint {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return 0
	}
	return uint(id)
}
