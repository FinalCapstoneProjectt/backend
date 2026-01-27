package proposals

import (
	"backend/internal/ai_checker"
	"backend/internal/auth"
	"backend/pkg/response"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service  *Service
	aiClient *ai_checker.Client
}

func NewHandler(s *Service, aiClient *ai_checker.Client) *Handler {
	return &Handler{service: s, aiClient: aiClient}
}

// DTOs
type SaveProposalRequest struct {
	TeamID           *uint  `json:"team_id"` // Optional
	Title            string `json:"title" binding:"required"`
	Abstract         string `json:"abstract"`
	ProblemStatement string `json:"problem_statement"`
	Objectives       string `json:"objectives"`
	Methodology      string `json:"methodology"`
	Timeline         string `json:"expected_timeline"`
	ExpectedOutcomes string `json:"expected_outcomes"`
}

type SubmitProposalRequest struct {
	TeamID uint `json:"team_id" binding:"required"`
}

// CreateProposal godoc
// @Summary Create a new proposal draft
// @Description Creates a new proposal ID with version 1. Team is optional at this stage.
// @Tags Proposals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param proposal body SaveProposalRequest true "Proposal details"
// @Success 201 {object} response.Response{data=domain.Proposal}
// @Router /proposals [post]
func (h *Handler) CreateProposal(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	var req SaveProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid inputs", err.Error())
		return
	}

	result, err := h.service.CreateDraft(h.mapRequestToInput(req), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create draft", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Draft created successfully", result)
}

// UpdateProposal godoc
// @Summary Update proposal or create revision
// @Description If Draft: updates existing. If Rejected/Revision: creates new version.
// @Tags Proposals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Param proposal body SaveProposalRequest true "Proposal details"
// @Success 200 {object} response.Response{data=domain.Proposal}
// @Router /proposals/{id} [put]
func (h *Handler) UpdateProposal(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	proposalID := parseID(c)
	if proposalID == 0 {
		return
	}

	var req SaveProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 👇 LOGGING ADDED
		fmt.Println("❌ JSON BIND ERROR:", err)
		response.Error(c, http.StatusBadRequest, "Invalid inputs", err.Error())
		return
	}

	// 👇 LOGGING ADDED
	fmt.Printf("📥 HANDLER RECEIVED: TeamID=%v, Title=%s\n", req.TeamID, req.Title)
	if req.TeamID != nil {
		fmt.Printf("   -> TeamID Value: %d\n", *req.TeamID)
	} else {
		fmt.Println("   -> TeamID is NIL")
	}

	result, err := h.service.UpdateProposal(proposalID, h.mapRequestToInput(req), claims.UserID)
	if err != nil {
		// Differentiate error types (400 vs 500) if needed
		response.Error(c, http.StatusBadRequest, "Failed to update proposal", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Proposal updated successfully", result)
}

// SubmitProposal godoc
// @Summary Submit proposal
// @Description Locks proposal and sends to Admin. Requires Finalized Team.
// @Tags Proposals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Param request body SubmitProposalRequest true "Team ID Confirmation"
// @Success 200 {object} response.Response
// @Router /proposals/{id}/submit [post]
func (h *Handler) SubmitProposal(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	proposalID := parseID(c)
	if proposalID == 0 {
		return
	}

	var req SubmitProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid inputs", err.Error())
		return
	}

	err := h.service.SubmitProposal(proposalID, req.TeamID, claims.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Submission failed", err.Error())
		return
	}

	data := gin.H{}
	if h.aiClient != nil {
		version, verErr := h.service.GetLatestVersion(proposalID)
		if verErr != nil {
			data["ai_error"] = verErr.Error()
		} else {
			aiResult, aiErr := h.aiClient.CheckProposalText(c.Request.Context(), ai_checker.ProposalCheckRequest{
				Title:      version.Title,
				Objectives: version.Objectives,
			})
			if aiErr != nil {
				data["ai_error"] = aiErr.Error()
			} else {
				data["ai_result"] = aiResult
			}
		}
	}

	if len(data) == 0 {
		response.JSON(c, http.StatusOK, "Proposal submitted successfully", nil)
		return
	}

	response.JSON(c, http.StatusOK, "Proposal submitted successfully", data)
}

// GET /proposals
// GetProposals godoc
// @Summary Get proposals
// @Description Retrieve proposals with optional filters
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param status query string false "Proposal status"
// @Param department_id query int false "Department ID"
// @Success 200 {object} response.Response{data=[]domain.Proposal}
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals [get]
func (h *Handler) GetProposals(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	status := c.Query("status")

	// Call service with user context from token
	proposals, err := h.service.GetProposals(
		status,
		claims.UserID,
		claims.Role,
		claims.DepartmentID,
	)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch proposals", err.Error())
		return
	}

	response.Success(c, proposals)
}

// GetProposal godoc
// @Summary Get proposal by ID
// @Description Retrieve a specific proposal by its ID
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response{data=domain.Proposal}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /proposals/{id} [get]
func (h *Handler) GetProposal(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		return
	}

	proposalID := parseID(c)
	if proposalID == 0 {
		return
	}

	proposal, err := h.service.GetProposal(
		proposalID,
		claims.UserID,
		claims.Role,
		claims.DepartmentID,
	)

	if err != nil {
		// Differentiate between Not Found and Forbidden
		if err.Error() == "proposal not found" {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
		} else {
			response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		return
	}

	response.Success(c, proposal)
}

// GetProposal godoc
// @Summary Get proposal by ID
// @Description Retrieve a specific proposal by its ID
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response{data=[]domain.ProposalVersion}
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id}/versions [get]
func (h *Handler) GetVersions(c *gin.Context) {
	id := parseID(c)
	if id == 0 {
		return
	}

	versions, err := h.service.GetVersions(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch versions", err.Error())
		return
	}

	response.Success(c, versions)
}

// DeleteProposal godoc
// @Summary Delete a proposal
// @Description Deletes a proposal if it is in Draft status
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /proposals/{id} [delete]
func (h *Handler) DeleteProposal(c *gin.Context) {
	id := parseID(c)
	if id == 0 {
		return
	}

	err := h.service.DeleteProposal(id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete proposal", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Proposal deleted successfully", nil)
}

// --- Helpers ---

func (h *Handler) mapRequestToInput(req SaveProposalRequest) ProposalInput {
	return ProposalInput{
		TeamID:           req.TeamID,
		Title:            req.Title,
		Abstract:         req.Abstract,
		ProblemStatement: req.ProblemStatement,
		Objectives:       req.Objectives,
		Methodology:      req.Methodology,
		Timeline:         req.Timeline,
		ExpectedOutcomes: req.ExpectedOutcomes,
	}
}

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

type AssignAdvisorRequest struct {
	AdvisorID uint `json:"advisor_id" binding:"required"`
}

// AssignAdvisor godoc
// @Summary Assign advisor to proposal
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /proposals/{id}/assign [patch]
func (h *Handler) AssignAdvisor(c *gin.Context) {
	id := parseID(c) // Helper
	var req AssignAdvisorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.AssignAdvisor(id, req.AdvisorID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Assignment failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, "Advisor assigned successfully", nil)
}
