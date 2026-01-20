package app

import (
	"backend/pkg/response"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *App) *gin.Engine {
	r := gin.Default()

	// Global Middlewares
	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(AuditMiddleware(app.AuditLogger))
	r.Use(RateLimitMiddleware())

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, "System is healthy", gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// API v1 Routes
	v1 := r.Group("/api/v1")
	{
		// Public Auth Routes
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", app.AuthHandler.Register)
			authRoutes.POST("/login", app.AuthHandler.Login)
			authRoutes.POST("/refresh", app.AuthHandler.RefreshToken)
		}

		// Protected Routes (require authentication)
		protected := v1.Group("")
		protected.Use(AuthMiddleware(app.Config))
		{
			// Auth Profile
			protected.GET("/auth/profile", app.AuthHandler.GetProfile)

			// Teams (Students)
			teams := protected.Group("/teams")
			{
				teams.POST("", RoleMiddleware("student"), app.TeamHandler.CreateTeam)
				teams.GET("", app.TeamHandler.GetTeams)
				teams.GET("/:id", app.TeamHandler.GetTeam)
				teams.GET("/:id/members", app.TeamHandler.GetTeamMembers)
				teams.POST("/:id/invite", RoleMiddleware("student"), app.TeamHandler.InviteMember)
				teams.POST("/:id/invitation/respond", RoleMiddleware("student"), app.TeamHandler.RespondToInvitation)
				teams.DELETE("/:id/members/:memberId", RoleMiddleware("student"), app.TeamHandler.RemoveMember)
			}

			// Proposals (Students & Teachers)
			proposals := protected.Group("/proposals")
			{
				proposals.POST("", RoleMiddleware("student"), app.ProposalHandler.CreateProposal)
				proposals.GET("", app.ProposalHandler.GetProposals)
				proposals.GET("/:id", app.ProposalHandler.GetProposal)
				proposals.POST("/:id/versions", RoleMiddleware("student"), app.ProposalHandler.CreateVersion)
				proposals.GET("/:id/versions", app.ProposalHandler.GetVersions)
				proposals.POST("/:id/submit", RoleMiddleware("student"), app.ProposalHandler.SubmitProposal)
				proposals.DELETE("/:id", RoleMiddleware("student"), app.ProposalHandler.DeleteProposal)
				proposals.GET("/:id/feedback", app.FeedbackHandler.GetProposalFeedback)
			}

			// Feedback (Teachers)
			feedback := protected.Group("/feedback")
			feedback.Use(RoleMiddleware("teacher"))
			{
				feedback.GET("/pending", app.FeedbackHandler.GetPendingProposals)
				feedback.POST("", app.FeedbackHandler.CreateFeedback)
				feedback.GET("/:id", app.FeedbackHandler.GetFeedback)
			}

			// Admin User Management
			admin := protected.Group("/admin")
			admin.Use(RoleMiddleware("admin"))
			{
				// User Management
				admin.POST("/users/teacher", app.UserHandler.CreateTeacher)
				admin.POST("/users/student", app.UserHandler.CreateStudent)
				admin.GET("/users", app.UserHandler.GetUsers)
				admin.GET("/users/:id", app.UserHandler.GetUser)
				admin.PATCH("/users/:id/status", app.UserHandler.UpdateUserStatus)
				admin.POST("/users/:id/assign-department", app.UserHandler.AssignDepartment)
				admin.DELETE("/users/:id", app.UserHandler.DeleteUser)
			}

			// Projects (Team creators can manage, all can view)
			projects := protected.Group("/projects")
			{
				projects.POST("", app.ProjectHandler.CreateProject)
				projects.GET("", app.ProjectHandler.GetProjects)
				projects.GET("/:id", app.ProjectHandler.GetProject)
				projects.PUT("/:id", app.ProjectHandler.UpdateProject)
				projects.POST("/:id/publish", app.ProjectHandler.PublishProject)
				//projects.GET("/:project_id/documentation", app.DocumentationHandler.GetProjectDocuments)
			}

			// Documentation (Team members upload, teachers review)
			//documentation := protected.Group("/documentation")
			//{
			//	documentation.POST("", app.DocumentationHandler.UploadDocument)
			//	documentation.GET("/:id", app.DocumentationHandler.GetDocument)
			//}

			// Documentation review (Teachers only)
			//docReview := protected.Group("/documentation")
			//docReview.Use(RoleMiddleware("teacher", "admin"))
			//{
			//	docReview.POST("/:id/review", app.DocumentationHandler.ReviewDocument)
			//}

			// Universities (Admin only)
			universities := protected.Group("/universities")
			universities.Use(RoleMiddleware("admin"))
			{
				universities.GET("", app.UniversityHandler.GetUniversities)
				universities.POST("", app.UniversityHandler.CreateUniversity)
				universities.GET("/:id", app.UniversityHandler.GetUniversity)
				universities.PUT("/:id", app.UniversityHandler.UpdateUniversity)
				universities.DELETE("/:id", app.UniversityHandler.DeleteUniversity)
			}

			// Departments (Admin/Department Head only)
			departments := protected.Group("/departments")
			departments.Use(RoleMiddleware("admin", "department_head"))
			{
				departments.GET("", app.DepartmentHandler.GetDepartments)
				departments.POST("", app.DepartmentHandler.CreateDepartment)
				departments.GET("/:id", app.DepartmentHandler.GetDepartment)
				departments.PUT("/:id", app.DepartmentHandler.UpdateDepartment)
				departments.DELETE("/:id", app.DepartmentHandler.DeleteDepartment)
			}
		}
	}

	return r
}
