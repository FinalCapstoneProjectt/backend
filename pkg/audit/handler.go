package audit

import (
	"backend/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles audit log API requests
type Handler struct {
	repo Repository
}

// NewHandler creates a new audit handler
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// GetAuditLogs returns audit logs with filtering and pagination
// @Summary Get audit logs
// @Description Get system audit logs with optional filters (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param entity_type query string false "Filter by entity type (proposal, team, user, etc.)"
// @Param entity_id query int false "Filter by specific entity ID"
// @Param actor_id query int false "Filter by actor user ID"
// @Param action query string false "Filter by action (create, submit, approve, etc.)"
// @Param from_date query string false "Start date (ISO 8601 format)"
// @Param to_date query string false "End date (ISO 8601 format)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/audit-logs [get]
func (h *Handler) GetAuditLogs(c *gin.Context) {
	filters := AuditFilters{}

	// Parse query parameters
	if entityType := c.Query("entity_type"); entityType != "" {
		filters.EntityType = entityType
	}

	if entityIDStr := c.Query("entity_id"); entityIDStr != "" {
		if id, err := strconv.ParseUint(entityIDStr, 10, 32); err == nil {
			filters.EntityID = uint(id)
		}
	}

	if actorIDStr := c.Query("actor_id"); actorIDStr != "" {
		if id, err := strconv.ParseUint(actorIDStr, 10, 32); err == nil {
			filters.ActorID = uint(id)
		}
	}

	if action := c.Query("action"); action != "" {
		filters.Action = action
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if t, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filters.FromDate = &t
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if t, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filters.ToDate = &t
		}
	}

	filters.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filters.Page = p
		}
	}

	filters.Limit = 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			filters.Limit = l
		}
	}

	logs, total, err := h.repo.GetLogs(filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch audit logs", err.Error())
		return
	}

	totalPages := (int(total) + filters.Limit - 1) / filters.Limit

	response.Success(c, gin.H{
		"audit_logs": logs,
		"pagination": gin.H{
			"page":        filters.Page,
			"limit":       filters.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetAuditLog returns a specific audit log entry
// @Summary Get audit log by ID
// @Description Get a specific audit log entry (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Audit Log ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /admin/audit-logs/{id} [get]
func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	log, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Audit log not found", err.Error())
		return
	}

	response.Success(c, log)
}
