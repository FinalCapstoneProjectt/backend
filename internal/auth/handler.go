package auth

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// Register routes
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.POST("/register", h.Register)
}

func (h *Handler) Login(c *gin.Context) {
	// TODO: Implement
}

func (h *Handler) Register(c *gin.Context) {
	// TODO: Implement
}
