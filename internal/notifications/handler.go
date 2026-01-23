package notifications

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

// GetNotifications godoc
// @Summary Get user notifications
// @Description Get all notifications for the current user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Notification}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications [get]
func (h *Handler) GetNotifications(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	notifications, err := h.service.GetUserNotifications(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch notifications", err.Error())
		return
	}

	response.Success(c, notifications)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]int64}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	count, err := h.service.GetUnreadCount(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get unread count", err.Error())
		return
	}

	response.Success(c, gin.H{"count": count})
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id}/read [patch]
func (h *Handler) MarkAsRead(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
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
		response.Error(c, http.StatusInternalServerError, "Failed to mark as read", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the current user as read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/read-all [patch]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	err := h.service.MarkAllAsRead(userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark all as read", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "All notifications marked as read", nil)
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete a specific notification
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/{id} [delete]
func (h *Handler) DeleteNotification(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "No authentication claims found")
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
		return
	}

	err = h.service.DeleteNotification(uint(id), userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete notification", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, "Notification deleted", nil)
}
