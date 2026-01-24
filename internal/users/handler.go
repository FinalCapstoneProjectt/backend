package users

import (
	"backend/pkg/response"
	"net/http"
	"strconv"
    "backend/internal/auth" // Ensure this is imported for TokenClaims

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// CreateTeacher godoc
// @Summary Register a new teacher
// @Description Admin registers or approves a new teacher account
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param teacher body CreateTeacherRequest true "Teacher registration details"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/teacher [post]
func (h *Handler) CreateTeacher(c *gin.Context) {
	var req CreateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.CreateTeacher(req)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Error(c, http.StatusConflict, "Email already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create teacher", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Teacher created successfully", user)
}

// CreateStudent godoc
// @Summary Register a new student
// @Description Admin registers or approves a new student account
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param student body CreateStudentRequest true "Student registration details"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/student [post]
func (h *Handler) CreateStudent(c *gin.Context) {
	var req CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.CreateStudent(req)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Error(c, http.StatusConflict, "Email already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create student", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, "Student created successfully", user)
}

// GetUsers godoc
// @Summary List all users
// @Description Admin retrieves list of users with optional filters
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param role query string false "Filter by role (admin, teacher, student)"
// @Param department_id query int false "Filter by department ID"
// @Param university_id query int false "Filter by university ID"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.Response{data=[]domain.User}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	role := c.Query("role")
	departmentIDStr := c.Query("department_id")
	universityIDStr := c.Query("university_id")
	isActiveStr := c.Query("is_active")

	var departmentID, universityID uint
	var isActive *bool

	if departmentIDStr != "" {
		id, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid department ID", err.Error())
			return
		}
		departmentID = uint(id)
	}

	if universityIDStr != "" {
		id, err := strconv.ParseUint(universityIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid university ID", err.Error())
			return
		}
		universityID = uint(id)
	}

	if isActiveStr != "" {
		active := isActiveStr == "true"
		isActive = &active
	}

	users, err := h.service.GetAllUsers(role, departmentID, universityID, isActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch users", err.Error())
		return
	}

	response.Success(c, users)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Admin retrieves user details by ID
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /admin/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.service.GetUser(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus godoc
// @Summary Activate or deactivate user
// @Description Admin controls user account activation status
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param status body UpdateUserStatusRequest true "User status"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{id}/status [patch]
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.service.UpdateUserStatus(uint(id), req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update user status", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "User status updated successfully", nil)
}

// AssignDepartment godoc
// @Summary Assign user to department
// @Description Admin assigns a teacher or student to a department
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param assignment body AssignDepartmentRequest true "Department assignment"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{id}/assign-department [post]
func (h *Handler) AssignDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req AssignDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.service.AssignDepartment(uint(id), req.DepartmentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to assign department", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Department assigned successfully", nil)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Admin deletes a user account (use with caution)
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	err = h.service.DeleteUser(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "User deleted successfully", nil)
}

// GetPeers godoc
// @Summary Get students in same department
// @Description Used for populating invite dropdowns
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Router /users/peers [get]
func (h *Handler) GetPeers(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	userClaims := claims.(*auth.TokenClaims)

	// 👇 FIXED: Use dynamic UniversityID from token
	users, err := h.service.GetPeers(userClaims.DepartmentID, userClaims.UniversityID, userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch peers", err.Error())
		return
	}
	
	for i := range users {
		users[i].Password = ""
	}

	response.Success(c, users)
}

// GetAdvisors godoc
// @Summary List advisors with workload
// @Description Admin sees list of advisors in their department with current team counts
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Router /admin/advisors [get]
func (h *Handler) GetAdvisors(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
        return
    }
    userClaims := claims.(*auth.TokenClaims)

    // Strict Data Isolation: Only get advisors from Admin's department
    data, err := h.service.GetDepartmentAdvisorsWithWorkload(userClaims.DepartmentID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to fetch advisors", err.Error())
        return
    }

    response.Success(c, data)
}

// GetDashboardStats godoc
// @Summary Get admin dashboard statistics
// @Description Aggregated stats for the Department Head dashboard
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Router /admin/stats [get]
func (h *Handler) GetDashboardStats(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	userClaims := claims.(*auth.TokenClaims)

	stats, err := h.service.GetAdminDashboardStats(userClaims.DepartmentID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch stats", err.Error())
		return
	}

	response.Success(c, stats)
}