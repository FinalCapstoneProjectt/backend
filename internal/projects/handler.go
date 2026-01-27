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

// GetPublicProjects godoc
// @Summary List all public projects
// @Description Get all public projects without authentication
// @Tags Projects
// @Produce json
// @Param department_id query int false "Filter by department ID"
// @Param year query int false "Filter by year"
// @Param search query string false "Search in title and summary"
// @Param sort query string false "Sort by: rating, date, views (default: rating)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} response.Response{data=[]domain.Project}
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/public [get]
func (h *Handler) GetPublicProjects(c *gin.Context) {
	filters := make(map[string]interface{})
	filters["visibility"] = "public"

	if deptID := c.Query("department_id"); deptID != "" {
		filters["department_id"] = deptID
	}
	if year := c.Query("year"); year != "" {
		filters["year"] = year
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if sort := c.Query("sort"); sort != "" {
		filters["sort"] = sort
	}

	// Pagination
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	filters["page"] = page
	filters["limit"] = limit

	projects, total, err := h.service.GetPublicProjects(filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch projects", err.Error())
		return
	}

	response.Success(c, gin.H{
		"projects": projects,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
	})
}

// GetPublicProject godoc
// @Summary Get public project by ID
// @Description Retrieve a public project without authentication
// @Tags Projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} response.Response{data=domain.Project}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /projects/public/{id} [get]
func (h *Handler) GetPublicProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	project, err := h.service.GetPublicProject(uint(id))
	if err != nil {
		if err.Error() == "project not found" {
			response.Error(c, http.StatusNotFound, "Project not found", nil)
			return
		}
		if err.Error() == "project is not public" {
			response.Error(c, http.StatusForbidden, "This project is not publicly accessible", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch project", err.Error())
		return
	}

	response.Success(c, project)
}

// IncrementShareCount godoc
// @Summary Increment project share count
// @Description Track when a project is shared
// @Tags Projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /projects/{id}/share [post]
func (h *Handler) IncrementShareCount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	newCount, err := h.service.IncrementShareCount(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update share count", err.Error())
		return
	}

	response.Success(c, gin.H{"share_count": newCount})
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

	project, err := h.service.UpdateProject(uint(id), req, userClaims.UserID, userClaims.Role)
	if err != nil {
		if err.Error() == "unauthorized: you cannot update this project" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
	response.JSON(c, http.StatusOK, "Project updated successfully", project)

}
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

	err = h.service.PublishProject(uint(id), userClaims.UserID, userClaims.Role)
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