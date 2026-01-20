package universities

import (
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

// CreateUniversity godoc
// @Summary Create a new university
// @Description Admin creates a new university with configuration settings
// @Tags Universities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param university body CreateUniversityRequest true "University details"
// @Success 201 {object} response.Response{data=domain.University}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /universities [post]
func (h *Handler) CreateUniversity(c *gin.Context) {
	var req CreateUniversityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	university, err := h.service.CreateUniversity(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create university", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "University created successfully", university)
}

// GetUniversities godoc
// @Summary Get all universities
// @Description Retrieve a list of all universities
// @Tags Universities
// @Produce json
// @Success 200 {object} response.Response{data=[]domain.University}
// @Failure 500 {object} response.ErrorResponse
// @Router /universities [get]
func (h *Handler) GetUniversities(c *gin.Context) {
	universities, err := h.service.GetAllUniversities()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch universities", err.Error())
		return
	}

	response.Success(c, universities)
}

// GetUniversity godoc
// @Summary Get university by ID
// @Description Retrieve a specific university by its ID
// @Tags Universities
// @Produce json
// @Param id path int true "University ID"
// @Success 200 {object} response.Response{data=domain.University}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /universities/{id} [get]
func (h *Handler) GetUniversity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid university ID", err.Error())
		return
	}

	university, err := h.service.GetUniversity(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "University not found", err.Error())
		return
	}

	response.Success(c, university)
}

// UpdateUniversity godoc
// @Summary Update university settings
// @Description Admin updates university configuration (academic year, visibility, AI checker)
// @Tags Universities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "University ID"
// @Param university body UpdateUniversityRequest true "University update data"
// @Success 200 {object} response.Response{data=domain.University}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /universities/{id} [put]
func (h *Handler) UpdateUniversity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid university ID", err.Error())
		return
	}

	var req UpdateUniversityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	university, err := h.service.UpdateUniversity(uint(id), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update university", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "University updated successfully", university)
}

// DeleteUniversity godoc
// @Summary Delete university
// @Description Admin deletes a university (use with caution)
// @Tags Universities
// @Produce json
// @Security BearerAuth
// @Param id path int true "University ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /universities/{id} [delete]
func (h *Handler) DeleteUniversity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid university ID", err.Error())
		return
	}

	err = h.service.DeleteUniversity(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete university", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "University deleted successfully", nil)
}
