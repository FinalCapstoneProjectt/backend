package files

import (
	"backend/internal/auth"
	"backend/pkg/enums"
	"backend/pkg/response"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// DownloadProposalFile godoc
// @Summary Download proposal document
// @Description Download a file from a proposal (access controlled)
// @Tags Files
// @Produce application/octet-stream
// @Security BearerAuth
// @Param proposal_id path int true "Proposal ID"
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /files/proposals/{proposal_id}/{filename} [get]
func (h *Handler) DownloadProposalFile(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	userClaims := claims.(*auth.TokenClaims)

	proposalID, err := strconv.ParseUint(c.Param("proposal_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid proposal ID", nil)
		return
	}

	filename := c.Param("filename")

	// Check access permission
	hasAccess, err := h.checkProposalAccess(uint(proposalID), userClaims)
	if err != nil || !hasAccess {
		response.Error(c, http.StatusForbidden, "You don't have access to this file", nil)
		return
	}

	// Construct file path
	filePath := filepath.Join("uploads", "proposals", strconv.FormatUint(proposalID, 10), filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "File not found", nil)
		return
	}

	// Serve file
	c.File(filePath)
}

// DownloadProjectFile godoc
// @Summary Download project document
// @Description Download a file from a project (public projects accessible to all)
// @Tags Files
// @Produce application/octet-stream
// @Param project_id path int true "Project ID"
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /files/projects/{project_id}/{filename} [get]
func (h *Handler) DownloadProjectFile(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("project_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	filename := c.Param("filename")

	// Check if project is public or user has access
	var project struct {
		Visibility string
		TeamID     uint
	}
	if err := h.db.Table("projects").Select("visibility, team_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Project not found", nil)
		return
	}

	// If project is private, check authentication
	if project.Visibility != "public" {
		claims, exists := c.Get("claims")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required for private projects", nil)
			return
		}
		userClaims := claims.(*auth.TokenClaims)

		hasAccess, _ := h.checkProjectAccess(uint(projectID), project.TeamID, userClaims)
		if !hasAccess {
			response.Error(c, http.StatusForbidden, "You don't have access to this file", nil)
			return
		}
	}

	// Construct file path
	filePath := filepath.Join("uploads", "projects", strconv.FormatUint(projectID, 10), filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "File not found", nil)
		return
	}

	// Serve file
	c.File(filePath)
}

// checkProposalAccess checks if user has access to a proposal
func (h *Handler) checkProposalAccess(proposalID uint, claims *auth.TokenClaims) (bool, error) {
	var proposal struct {
		TeamID    *uint
		AdvisorID *uint
		CreatedBy uint
	}

	if err := h.db.Table("proposals").Select("team_id, advisor_id, created_by").Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return false, err
	}

	// Admin can access proposals in their department
	if claims.Role == enums.RoleAdmin {
		return true, nil
	}

	// Advisor can access assigned proposals
	if claims.Role == enums.RoleAdvisor && proposal.AdvisorID != nil && *proposal.AdvisorID == claims.UserID {
		return true, nil
	}

	// Creator can access
	if proposal.CreatedBy == claims.UserID {
		return true, nil
	}

	// Team member can access
	if proposal.TeamID != nil {
		var count int64
		h.db.Table("team_members").Where("team_id = ? AND user_id = ?", *proposal.TeamID, claims.UserID).Count(&count)
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

// checkProjectAccess checks if user has access to a private project
func (h *Handler) checkProjectAccess(projectID uint, teamID uint, claims *auth.TokenClaims) (bool, error) {
	// Admin can access
	if claims.Role == enums.RoleAdmin {
		return true, nil
	}

	// Team member can access
	var count int64
	h.db.Table("team_members").Where("team_id = ? AND user_id = ?", teamID, claims.UserID).Count(&count)
	if count > 0 {
		return true, nil
	}

	return false, nil
}
