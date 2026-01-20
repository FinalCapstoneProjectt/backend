package projects

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

// CreateProject godoc
// @Summary Create project from approved proposal
// @Description Convert an approved proposal into a formal project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project body CreateProjectRequest true "Project details"
// @Success 201 {object} response.Response{data=domain.Project}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	project, err := h.service.CreateProject(req, userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create project", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Project created successfully", project)
}

// GetProjects godoc
// @Summary List all projects
// @Description Get all projects with optional filters
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param visibility query string false "Filter by visibility (private, public)"
// @Param department_id query int false "Filter by department ID"
// @Param team_id query int false "Filter by team ID"
// @Success 200 {object} response.Response{data=[]domain.Project}
// @Failure 500 {object} response.ErrorResponse
// @Router /projects [get]
func (h *Handler) GetProjects(c *gin.Context) {
	filters := make(map[string]interface{})

	if visibility := c.Query("visibility"); visibility != "" {
		filters["visibility"] = visibility
	}
	if deptID := c.Query("department_id"); deptID != "" {
		filters["department_id"] = deptID
	}
	if teamID := c.Query("team_id"); teamID != "" {
		filters["team_id"] = teamID
	}

	projects, err := h.service.GetProjects(filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch projects", err.Error())
		return
	}

	response.Success(c, projects)
}

// GetProject godoc
// @Summary Get project by ID
// @Description Retrieve specific project details
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} response.Response{data=domain.Project}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /projects/{id} [get]
func (h *Handler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	project, err := h.service.GetProject(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Project not found", err.Error())
		return
	}

	response.Success(c, project)
}

// UpdateProject godoc
// @Summary Update project details
// @Description Update project summary and keywords
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param project body UpdateProjectRequest true "Updated project details"
// @Success 200 {object} response.Response{data=domain.Project}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	project, err := h.service.UpdateProject(uint(id), req, userClaims.UserID)
	if err != nil {
		if err.Error() == "only team creator can update project" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update project", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Project updated successfully", project)
}

// PublishProject godoc
// @Summary Publish project to public archive
// @Description Make project visible to public users
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id}/publish [post]
func (h *Handler) PublishProject(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	err = h.service.PublishProject(uint(id), userClaims.UserID)
	if err != nil {
		if err.Error() == "only team creator can publish project" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to publish project", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Project published successfully", nil)
}
