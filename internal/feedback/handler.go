package feedback

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

// GetPendingProposals godoc
// @Summary Get pending proposals for review
// @Description Teacher gets all proposals awaiting their review
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Proposal}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /feedback/pending [get]
func (h *Handler) GetPendingProposals(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*auth.TokenClaims)

	proposals, err := h.service.GetPendingProposals(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Fetch failed", err.Error())
		return
	}
	response.Success(c, proposals)
}


// CreateFeedback godoc
// @Summary Submit feedback for a proposal
// @Description Teacher reviews proposal and submits feedback (approve, revise, reject)
// @Tags Feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param feedback body CreateFeedbackRequest true "Feedback details"
// @Success 201 {object} response.Response{data=domain.Feedback}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /feedback [post]
func (h *Handler) CreateFeedback(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*auth.TokenClaims)

	var req CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	feedback, err := h.service.CreateFeedback(req, userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.JSON(c, http.StatusCreated, "Feedback submitted", feedback)
}

// GetProposalFeedback godoc
// @Summary Get all feedback for a proposal
// @Description Retrieve all feedback history for a specific proposal
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response{data=[]domain.Feedback}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id}/feedback [get]
func (h *Handler) GetProposalFeedback(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", err.Error())
		return
	}

	feedbacks, err := h.service.GetProposalFeedback(uint(id), userClaims.UserID)
	if err != nil {
		if err.Error() == "you don't have permission to view this feedback" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch feedback", err.Error())
		return
	}

	response.Success(c, feedbacks)
}

// GetFeedback godoc
// @Summary Get feedback by ID
// @Description Retrieve specific feedback details
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feedback ID"
// @Success 200 {object} response.Response{data=domain.Feedback}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /feedback/{id} [get]
func (h *Handler) GetFeedback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid feedback ID", err.Error())
		return
	}

	feedback, err := h.service.GetFeedbackByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Feedback not found", err.Error())
		return
	}

	response.Success(c, feedback)
}

