package notifications

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles notification API requests
type Handler struct {
	service *Service
}

// NewHandler creates a new notification handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetNotifications returns notifications for the authenticated user
// @Summary Get user notifications
// @Description Get all notifications for the authenticated user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param is_read query bool false "Filter by read status"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 50)"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [get]
func (h *Handler) GetNotifications(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	// Parse query parameters
	var isRead *bool
	if isReadStr := c.Query("is_read"); isReadStr != "" {
		val := isReadStr == "true"
		isRead = &val
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	notifications, unreadCount, err := h.service.GetUserNotifications(userClaims.UserID, isRead, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch notifications", err.Error())
		return
	}

	response.Success(c, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
		},
	})
}

// MarkAsRead marks a notification as read
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /notifications/{id}/mark-read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.service.MarkAsRead(uint(id), userClaims.UserID)
	if err != nil {
		if err.Error() == "notification not found" || err.Error() == "notification does not belong to user" {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications as read
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/mark-all-read [post]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	err := h.service.MarkAllAsRead(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "All notifications marked as read", nil)
}

// GetUnreadCount returns the count of unread notifications
// @Summary Get unread notification count
// @Description Get the count of unread notifications for the authenticated user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	count, err := h.service.GetUnreadCount(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get unread count", err.Error())
		return
	}

	response.Success(c, gin.H{
		"unread_count": count,
	})
}
