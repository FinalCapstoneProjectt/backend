package proposals

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func saveUploadedFile(c *gin.Context, formField string) (string, string, int64, error) {
	fileHeader, err := c.FormFile(formField)
	if err != nil {
		return "", "", 0, err
	}

	uploadDir := filepath.Join("uploads", "proposals")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", "", 0, err
	}

	filename := fmt.Sprintf("%s_%s", uuid.New().String(), fileHeader.Filename)
	filePath := filepath.Join(uploadDir, filename)

	src, err := fileHeader.Open()
	if err != nil {
		return "", "", 0, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", 0, err
	}
	defer dst.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)
	bytesWritten, err := io.Copy(writer, src)
	if err != nil {
		return "", "", 0, err
	}

	fileURL := "/uploads/proposals/" + filename
	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))
	return fileURL, fileHash, bytesWritten, nil
}

// CreateProposal godoc
// @Summary Create a new proposal (draft)
// @Description Student creates a new proposal in draft state
// @Tags Proposals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param proposal body CreateProposalRequest true "Proposal details"
// @Success 201 {object} response.Response{data=domain.Proposal}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals [post]
func (h *Handler) CreateProposal(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	if strings.Contains(c.ContentType(), "multipart/form-data") {
		teamIDStr := c.PostForm("team_id")
		if teamIDStr == "" {
			response.Error(c, http.StatusBadRequest, "Invalid request body", "team_id is required")
			return
		}
		teamIDParsed, err := strconv.ParseUint(teamIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid team_id", err.Error())
			return
		}

		proposal, err := h.service.CreateProposal(CreateProposalRequest{TeamID: uint(teamIDParsed)})
		if err != nil {
			if err.Error() == "team already has a proposal" {
				response.Error(c, http.StatusConflict, "Team already has a proposal", err.Error())
				return
			}
			response.Error(c, http.StatusInternalServerError, "Failed to create proposal", err.Error())
			return
		}

		fileURL, fileHash, fileSize, err := saveUploadedFile(c, "file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid file upload", err.Error())
			return
		}

		versionReq := CreateVersionRequest{
			Title:            c.PostForm("title"),
			Objectives:       c.PostForm("objectives"),
			Methodology:      c.PostForm("methodology"),
			ExpectedOutcomes: c.PostForm("expected_outcomes"),
			FileURL:          fileURL,
			FileHash:         fileHash,
			FileSizeBytes:    fileSize,
			IPAddress:        c.ClientIP(),
			UserAgent:        c.GetHeader("User-Agent"),
		}
		if requestID, ok := c.Get("request_id"); ok {
			if requestIDStr, ok := requestID.(string); ok {
				versionReq.SessionID = requestIDStr
			}
		}

		if versionReq.Title == "" || versionReq.Objectives == "" || versionReq.Methodology == "" || versionReq.ExpectedOutcomes == "" {
			response.Error(c, http.StatusBadRequest, "Invalid request body", "title, objectives, methodology, and expected_outcomes are required")
			return
		}

		_, err = h.service.CreateVersion(proposal.ID, versionReq, userClaims.UserID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to create proposal version", err.Error())
			return
		}

		updated, err := h.service.GetProposal(proposal.ID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to load proposal", err.Error())
			return
		}

		response.JSON(c, http.StatusCreated, "Proposal created successfully", updated)
		return
	}

	var req CreateProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	proposal, err := h.service.CreateProposal(req)
	if err != nil {
		if err.Error() == "team already has a proposal" {
			response.Error(c, http.StatusConflict, "Team already has a proposal", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create proposal", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Proposal created successfully", proposal)
}

// GetProposals godoc
// @Summary List proposals
// @Description Get all proposals (filtered by status, department)
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (draft, submitted, under_review, etc.)"
// @Param department_id query int false "Filter by department ID"
// @Success 200 {object} response.Response{data=[]domain.Proposal}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals [get]
func (h *Handler) GetProposals(c *gin.Context) {
	status := c.Query("status")
	departmentIDStr := c.Query("department_id")

	var departmentID uint
	if departmentIDStr != "" {
		id, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid department ID", err.Error())
			return
		}
		departmentID = uint(id)
	}

	proposals, err := h.service.GetProposals(status, departmentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch proposals", err.Error())
		return
	}

	response.Success(c, proposals)
}

// GetProposal godoc
// @Summary Get proposal by ID
// @Description Retrieve proposal details with versions and feedback
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", err.Error())
		return
	}

	proposal, err := h.service.GetProposal(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Proposal not found", err.Error())
		return
	}

	response.Success(c, proposal)
}

// CreateVersion godoc
// @Summary Create a new proposal version
// @Description Add a new version to a draft or revision-required proposal
// @Tags Proposals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Param version body CreateVersionRequest true "Version details"
// @Success 201 {object} response.Response{data=domain.ProposalVersion}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id}/versions [post]
func (h *Handler) CreateVersion(c *gin.Context) {
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

	var req CreateVersionRequest
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		fileURL, fileHash, fileSize, err := saveUploadedFile(c, "file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid file upload", err.Error())
			return
		}

		req = CreateVersionRequest{
			Title:            c.PostForm("title"),
			Objectives:       c.PostForm("objectives"),
			Methodology:      c.PostForm("methodology"),
			ExpectedOutcomes: c.PostForm("expected_outcomes"),
			FileURL:          fileURL,
			FileHash:         fileHash,
			FileSizeBytes:    fileSize,
			IPAddress:        c.ClientIP(),
			UserAgent:        c.GetHeader("User-Agent"),
		}
		if requestID, ok := c.Get("request_id"); ok {
			if requestIDStr, ok := requestID.(string); ok {
				req.SessionID = requestIDStr
			}
		}

		if req.Title == "" || req.Objectives == "" || req.Methodology == "" || req.ExpectedOutcomes == "" {
			response.Error(c, http.StatusBadRequest, "Invalid request body", "title, objectives, methodology, and expected_outcomes are required")
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}
		req.IPAddress = c.ClientIP()
		req.UserAgent = c.GetHeader("User-Agent")
		if requestID, ok := c.Get("request_id"); ok {
			if requestIDStr, ok := requestID.(string); ok {
				req.SessionID = requestIDStr
			}
		}
	}

	version, err := h.service.CreateVersion(uint(id), req, userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create version", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Version created successfully", version)
}

// GetVersions godoc
// @Summary Get all versions of a proposal
// @Description Retrieve version history for a proposal
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response{data=[]domain.ProposalVersion}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id}/versions [get]
func (h *Handler) GetVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", err.Error())
		return
	}

	versions, err := h.service.GetVersions(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch versions", err.Error())
		return
	}

	response.Success(c, versions)
}

// SubmitProposal godoc
// @Summary Submit proposal for review
// @Description Team leader submits proposal, locks it, and notifies teacher
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id}/submit [post]
func (h *Handler) SubmitProposal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", err.Error())
		return
	}

	err = h.service.SubmitProposal(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit proposal", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Proposal submitted successfully", nil)
}

// DeleteProposal godoc
// @Summary Delete a draft proposal
// @Description Delete a proposal (only allowed for drafts)
// @Tags Proposals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Proposal ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /proposals/{id} [delete]
func (h *Handler) DeleteProposal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", err.Error())
		return
	}

	err = h.service.DeleteProposal(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete proposal", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Proposal deleted successfully", nil)
}
