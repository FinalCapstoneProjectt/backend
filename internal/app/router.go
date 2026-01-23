package app

import (
	"backend/pkg/enums"
	"backend/pkg/response"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *App) *gin.Engine {
	r := gin.Default()

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

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

			// User Search (available to all authenticated users)
			users := protected.Group("/users")
			{
				users.GET("/students/search", app.UserHandler.SearchStudents)
				users.GET("/teachers", app.UserHandler.GetTeachers)
			}

			// Teams (Students)
			teams := protected.Group("/teams")
			{
				teams.POST("", RoleMiddleware(enums.RoleStudent), app.TeamHandler.CreateTeam)
				teams.GET("", app.TeamHandler.GetTeams)
				teams.GET("/:id", app.TeamHandler.GetTeam)
				teams.GET("/:id/members", app.TeamHandler.GetTeamMembers)
				teams.POST("/:id/invite", RoleMiddleware(enums.RoleStudent), app.TeamHandler.InviteMember)
				teams.POST("/:id/invitation/respond", RoleMiddleware(enums.RoleStudent), app.TeamHandler.RespondToInvitation)
				teams.POST("/:id/approval", RoleMiddleware(enums.RoleTeacher), app.TeamHandler.ApproveTeam)
				teams.DELETE("/:id/members/:memberId", RoleMiddleware(enums.RoleStudent), app.TeamHandler.RemoveMember)
			}

			// Proposals (Students & Teachers)
			proposals := protected.Group("/proposals")
			{
				proposals.POST("", RoleMiddleware(enums.RoleStudent), app.ProposalHandler.CreateProposal)
				proposals.GET("", app.ProposalHandler.GetProposals)
				proposals.GET("/:id", app.ProposalHandler.GetProposal)
				proposals.POST("/:id/versions", RoleMiddleware(enums.RoleStudent), app.ProposalHandler.CreateVersion)
				proposals.GET("/:id/versions", app.ProposalHandler.GetVersions)
				proposals.POST("/:id/submit", RoleMiddleware(enums.RoleStudent), app.ProposalHandler.SubmitProposal)
				proposals.DELETE("/:id", RoleMiddleware(enums.RoleStudent), app.ProposalHandler.DeleteProposal)
				proposals.GET("/:id/feedback", app.FeedbackHandler.GetProposalFeedback)
			}

			// Feedback (Teachers)
			feedback := protected.Group("/feedback")
			feedback.Use(RoleMiddleware(enums.RoleTeacher))
			{
				feedback.GET("/pending", app.FeedbackHandler.GetPendingProposals)
				feedback.POST("", app.FeedbackHandler.CreateFeedback)
				feedback.GET("/:id", app.FeedbackHandler.GetFeedback)
			}

			// Admin User Management
			admin := protected.Group("/admin")
			admin.Use(RoleMiddleware(enums.RoleAdmin))
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

			// Notifications
			notificationsRoutes := protected.Group("/notifications")
			{
				notificationsRoutes.GET("", app.NotificationHandler.GetNotifications)
				notificationsRoutes.GET("/unread-count", app.NotificationHandler.GetUnreadCount)
				notificationsRoutes.PATCH("/:id/read", app.NotificationHandler.MarkAsRead)
				notificationsRoutes.PATCH("/read-all", app.NotificationHandler.MarkAllAsRead)
				notificationsRoutes.DELETE("/:id", app.NotificationHandler.DeleteNotification)
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
			universities.Use(RoleMiddleware(enums.RoleAdmin))
			{
				universities.GET("", app.UniversityHandler.GetUniversities)
				universities.POST("", app.UniversityHandler.CreateUniversity)
				universities.GET("/:id", app.UniversityHandler.GetUniversity)
				universities.PUT("/:id", app.UniversityHandler.UpdateUniversity)
				universities.DELETE("/:id", app.UniversityHandler.DeleteUniversity)
			}

			// Departments (Admin/Department Head only)
			departments := protected.Group("/departments")
			departments.Use(RoleMiddleware(enums.RoleAdmin, "department_head"))
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
