package reviews

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles project review API requests
type Handler struct {
	service *Service
}

// NewHandler creates a new review handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateReviewRequest represents the request body for creating a review
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=500"`
}

// CreateReview creates a new review for a project
// @Summary Add project review
// @Description Add a review and rating to a public project
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param review body CreateReviewRequest true "Review details"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /projects/{id}/reviews [post]
func (h *Handler) CreateReview(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims := claims.(*auth.TokenClaims)

	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	review, avgRating, err := h.service.CreateReview(userClaims.UserID, uint(projectID), req.Rating, req.Comment)
	if err != nil {
		switch err.Error() {
		case "project not found":
			response.Error(c, http.StatusNotFound, err.Error(), nil)
		case "can only review public projects":
			response.Error(c, http.StatusForbidden, err.Error(), nil)
		case "you have already reviewed this project":
			response.Error(c, http.StatusConflict, err.Error(), nil)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to create review", err.Error())
		}
		return
	}

	response.JSON(c, http.StatusCreated, "Review submitted", gin.H{
		"review":          review,
		"updated_average": avgRating,
	})
}

// GetProjectReviews returns all reviews for a project
// @Summary Get project reviews
// @Description Get all reviews and average rating for a project
// @Tags Reviews
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /projects/{id}/reviews [get]
func (h *Handler) GetProjectReviews(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	reviews, avgRating, err := h.service.GetProjectReviews(uint(projectID))
	if err != nil {
		if err.Error() == "project not found" {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch reviews", err.Error())
		return
	}

	response.Success(c, gin.H{
		"reviews":        reviews,
		"average_rating": avgRating,
		"total_reviews":  len(reviews),
	})
}
