package ai_checker

import (
	"backend/pkg/response"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	client *Client
}

type ProposalCheckRequest struct {
	Title      string `json:"title" binding:"required" example:"Project Title"`
	Objectives string `json:"objectives" binding:"required" example:"Project objectives text"`
}

type SyncProject struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

// HealthCheck godoc
// @Summary AI service health check
// @Description Checks if the AI service is reachable
// @Tags AI Checker
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 503 {object} response.ErrorResponse
// @Router /ai-checker/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	if err := h.client.Health(c.Request.Context()); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "AI service unavailable", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "AI service available", gin.H{"status": "ok"})
}

// CheckProposalText godoc
// @Summary Analyze proposal text
// @Description Sends proposal title and objectives to the AI service
// @Tags AI Checker
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body ProposalCheckRequest true "Proposal content"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 502 {object} response.ErrorResponse
// @Router /ai-checker/proposal-check [post]
func (h *Handler) CheckProposalText(c *gin.Context) {
	var req ProposalCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.client.CheckProposalText(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "AI service request failed", err.Error())
		return
	}

	response.Success(c, result)
}

// CheckProposalFile godoc
// @Summary Analyze proposal file
// @Description Uploads proposal PDF/DOCX file to the AI service
// @Tags AI Checker
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Proposal file"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 502 {object} response.ErrorResponse
// @Router /ai-checker/proposal-check-file [post]
func (h *Handler) CheckProposalFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File is required", err.Error())
		return
	}

	opened, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to read file", err.Error())
		return
	}
	defer opened.Close()

	content, err := io.ReadAll(opened)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to read file", err.Error())
		return
	}

	result, err := h.client.CheckProposalFile(c.Request.Context(), file.Filename, content)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "AI service request failed", err.Error())
		return
	}

	response.Success(c, result)
}
