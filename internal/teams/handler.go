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
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	team, err := h.service.CreateTeam(req, userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create team", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Team created successfully", team)
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
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	teams, err := h.service.GetMyTeams(userClaims.UserID)
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

// RemoveMember godoc
// @Summary Remove a member from team
// @Description Team leader removes a member from the team
// @Tags Teams
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param memberId path int true "Member User ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams/{id}/members/{memberId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
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

	memberID, err := strconv.ParseUint(c.Param("memberId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid member ID", err.Error())
		return
	}

	err = h.service.RemoveMember(uint(id), uint(memberID), userClaims.UserID)
	if err != nil {
		if err.Error() == "only team leader can remove members" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Member removed successfully", nil)
}

// ApproveTeam godoc
// @Summary Approve or reject a team
// @Description Advisor approves or rejects a team assigned to them
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param body body ApproveTeamRequest true "Approve or reject"
// @Success 200 {object} response.Response{data=domain.Team}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /teams/{id}/approval [post]
func (h *Handler) ApproveTeam(c *gin.Context) {
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

	var req ApproveTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	team, err := h.service.ApproveTeam(uint(id), userClaims.UserID, req.Approve)
	if err != nil {
		if err.Error() == "only assigned advisor can approve this team" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		if err.Error() == "team is not pending advisor approval" {
			response.Error(c, http.StatusBadRequest, "Invalid team state", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update team status", err.Error())
		return
	}

	message := "Team rejected"
	if req.Approve {
		message = "Team approved"
	}
	response.JSON(c, http.StatusOK, message, team)
}
