package departments

import (
	"backend/internal/domain"
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

// CreateDepartment godoc
// @Summary Create a new department
// @Description Admin creates a new department under a university
// @Tags Departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param department body CreateDepartmentRequest true "Department details"
// @Success 201 {object} response.Response{data=domain.Department}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /departments [post]
func (h *Handler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	department, err := h.service.CreateDepartment(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create department", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Department created successfully", department)
}

// GetDepartments godoc
// @Summary Get all departments
// @Description Retrieve a list of all departments (optionally filtered by university)
// @Tags Departments
// @Produce json
// @Param university_id query int false "Filter by university ID"
// @Success 200 {object} response.Response{data=[]domain.Department}
// @Failure 500 {object} response.ErrorResponse
// @Router /departments [get]
func (h *Handler) GetDepartments(c *gin.Context) {
	universityIDStr := c.Query("university_id")

	var departments []domain.Department
	var err error

	if universityIDStr != "" {
		universityID, parseErr := strconv.ParseUint(universityIDStr, 10, 32)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid university ID", parseErr.Error())
			return
		}
		departments, err = h.service.GetDepartmentsByUniversity(uint(universityID))
	} else {
		departments, err = h.service.GetAllDepartments()
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch departments", err.Error())
		return
	}

	response.Success(c, departments)
}

// GetDepartment godoc
// @Summary Get department by ID
// @Description Retrieve a specific department by its ID
// @Tags Departments
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} response.Response{data=domain.Department}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /departments/{id} [get]
func (h *Handler) GetDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid department ID", err.Error())
		return
	}

	department, err := h.service.GetDepartment(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Department not found", err.Error())
		return
	}

	response.Success(c, department)
}

// UpdateDepartment godoc
// @Summary Update department
// @Description Admin updates department information
// @Tags Departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Param department body UpdateDepartmentRequest true "Department update data"
// @Success 200 {object} response.Response{data=domain.Department}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /departments/{id} [put]
func (h *Handler) UpdateDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid department ID", err.Error())
		return
	}

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	department, err := h.service.UpdateDepartment(uint(id), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update department", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Department updated successfully", department)
}

// DeleteDepartment godoc
// @Summary Delete department
// @Description Admin deletes a department (use with caution)
// @Tags Departments
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /departments/{id} [delete]
func (h *Handler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid department ID", err.Error())
		return
	}

	err = h.service.DeleteDepartment(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete department", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Department deleted successfully", nil)
}
