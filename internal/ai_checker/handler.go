package ai_checker

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles AI checker API requests
type Handler struct {
	client       *Client
	proposalRepo ProposalRepository
}

// ProposalRepository interface for accessing proposals
type ProposalRepository interface {
	GetByID(id uint) (ProposalData, error)
}

// ProposalData contains proposal information needed for AI analysis
type ProposalData struct {
	ID         uint
	Title      string
	Objectives string
	TeamID     *uint
	AdvisorID  *uint
}

// NewHandler creates a new AI handler
func NewHandler(client *Client, proposalRepo ProposalRepository) *Handler {
	return &Handler{
		client:       client,
		proposalRepo: proposalRepo,
	}
}

// AnalyzeProposalRequest is the request body for proposal analysis
type AnalyzeProposalRequest struct {
	ProposalID uint `json:"proposal_id" binding:"required"`
	VersionID  uint `json:"version_id"`
	Title      string `json:"title"`
	Objectives string `json:"objectives"`
}

// AnalyzeProposal analyzes a proposal using the AI service
// @Summary Analyze a proposal with AI
// @Description Get AI-powered analysis of a proposal including summary, structure hints, and similarity warnings
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AnalyzeProposalRequest true "Analysis request"
// @Success 200 {object} response.Response{data=ProposalCheckResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /ai/analyze-proposal [post]
func (h *Handler) AnalyzeProposal(c *gin.Context) {
	var req AnalyzeProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Use provided title/objectives or fetch from proposal
	title := req.Title
	objectives := req.Objectives

	if title == "" || objectives == "" {
		// Fetch from proposal if not provided directly
		if h.proposalRepo != nil && req.ProposalID > 0 {
			proposal, err := h.proposalRepo.GetByID(req.ProposalID)
			if err != nil {
				response.Error(c, http.StatusNotFound, "Proposal not found", err.Error())
				return
			}
			if title == "" {
				title = proposal.Title
			}
			if objectives == "" {
				objectives = proposal.Objectives
			}
		}
	}

	if title == "" || objectives == "" {
		response.Error(c, http.StatusBadRequest, "Title and objectives are required", nil)
		return
	}

	// Call AI service
	result, err := h.client.CheckProposal(title, objectives)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "AI service unavailable", err.Error())
		return
	}

	response.Success(c, result)
}

// CheckSimilarityRequest is the request for similarity check
type CheckSimilarityRequest struct {
	ProposalID uint    `form:"proposal_id" binding:"required"`
	Threshold  float64 `form:"threshold"`
}

// CheckSimilarity checks proposal similarity against existing projects
// @Summary Check proposal similarity
// @Description Check if a proposal is similar to existing approved projects
// @Tags AI
// @Produce json
// @Security BearerAuth
// @Param proposal_id query int true "Proposal ID"
// @Param threshold query number false "Similarity threshold (0.0-1.0)"
// @Success 200 {object} response.Response{data=[]SimilarityWarning}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /ai/check-similarity [get]
func (h *Handler) CheckSimilarity(c *gin.Context) {
	// Get claims to verify user is advisor
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	_ = claims.(*auth.TokenClaims)

	proposalIDStr := c.Query("proposal_id")
	if proposalIDStr == "" {
		response.Error(c, http.StatusBadRequest, "proposal_id is required", nil)
		return
	}

	proposalID, err := strconv.ParseUint(proposalIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal_id", err.Error())
		return
	}

	// Fetch proposal data
	if h.proposalRepo == nil {
		response.Error(c, http.StatusInternalServerError, "Proposal repository not configured", nil)
		return
	}

	proposal, err := h.proposalRepo.GetByID(uint(proposalID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Proposal not found", err.Error())
		return
	}

	// Call AI service for analysis (includes similarity)
	result, err := h.client.CheckProposal(proposal.Title, proposal.Objectives)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "AI service unavailable", err.Error())
		return
	}

	// Filter by threshold if provided
	threshold := 0.3 // default
	if thresholdStr := c.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil && t >= 0 && t <= 1 {
			threshold = t
		}
	}

	// Filter warnings by threshold
	var filteredWarnings []SimilarityWarning
	for _, w := range result.SimilarityWarnings {
		if w.SimilarityScore >= threshold {
			filteredWarnings = append(filteredWarnings, w)
		}
	}

	response.Success(c, gin.H{
		"similar_projects": filteredWarnings,
		"threshold":        threshold,
		"checked_at":       c.Request.Header.Get("Date"),
	})
}

// HealthCheck checks AI service health
// @Summary Check AI service health
// @Description Verify the AI service is available
// @Tags AI
// @Produce json
// @Success 200 {object} response.Response
// @Failure 503 {object} response.ErrorResponse
// @Router /ai/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	if err := h.client.HealthCheck(); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "AI service unavailable", err.Error())
		return
	}

	response.Success(c, gin.H{
		"status":  "ok",
		"service": "ai-checker",
	})
}
