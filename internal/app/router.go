package app

import (
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS and other global middlewares would go here

	r.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, "System is healthy", nil)
	})

	v1 := r.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) { response.JSON(c, http.StatusOK, "Login Mock", nil) })
			auth.POST("/register", func(c *gin.Context) { response.JSON(c, http.StatusOK, "Register Mock", nil) })
		}

		// Team routes (Protected)
		teams := v1.Group("/teams")
		teams.Use(AuthMiddleware())
		{
			teams.GET("", func(c *gin.Context) { response.JSON(c, http.StatusOK, "Get Teams Mock", nil) })
		}

		// Proposal routes (Protected)
		proposals := v1.Group("/proposals")
		proposals.Use(AuthMiddleware())
		{
			proposals.GET("", func(c *gin.Context) { response.JSON(c, http.StatusOK, "Get Proposals Mock", nil) })
		}
	}

	return r
}
