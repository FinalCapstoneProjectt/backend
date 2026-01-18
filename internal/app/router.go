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
		// ============ PUBLIC ROUTES (No Auth Required) ============
		
		// Universities
		universities := v1.Group("/universities")
		{
			universities.GET("", app.UniversityHandler.GetUniversities)
			universities.GET("/:id", app.UniversityHandler.GetUniversity)
		}

		// Departments
		departments := v1.Group("/departments")
		{
			departments.GET("", app.DepartmentHandler.GetDepartments)
			departments.GET("/:id", app.DepartmentHandler.GetDepartment)
		}

		// Public Auth Routes
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", app.AuthHandler.Register)
			authRoutes.POST("/login", app.AuthHandler.Login)
			authRoutes.POST("/refresh", app.AuthHandler.RefreshToken)
			authRoutes.POST("/forgot-password", app.AuthHandler.ForgotPassword)
			authRoutes.POST("/reset-password", app.AuthHandler.ResetPassword)
		}

		// Public Project Routes (No Auth Required)
		publicProjects := v1.Group("/projects/public")
		{
			publicProjects.GET("", app.ProjectHandler.GetPublicProjects)
			publicProjects.GET("/:id", app.ProjectHandler.GetPublicProject)
		}

		// Public File Downloads (for public projects)
		v1.GET("/files/projects/:project_id/*filename", app.FileHandler.DownloadProjectFile)

		
		// ============ PROTECTED ROUTES (Auth Required) ============
		protected := v1.Group("")
		protected.Use(AuthMiddleware(app.Config))
		{
			// Auth Profile & Password Management
			protected.GET("/auth/profile", app.AuthHandler.GetProfile)
			protected.PUT("/auth/profile", app.AuthHandler.UpdateProfile)
			protected.POST("/auth/change-password", app.AuthHandler.ChangePassword)

			// Peer List for Invites
			protected.GET("/users/peers", app.UserHandler.GetPeers)

			// Secure File Downloads (for proposals)
			protected.GET("/files/proposals/:proposal_id/*filename", app.FileHandler.DownloadProposalFile)

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
				teams.POST("/:id/transfer-leadership", RoleMiddleware("student"), app.TeamHandler.TransferLeadership)
				teams.DELETE("/:id", RoleMiddleware("student"), app.TeamHandler.DeleteTeam)
				teams.POST("/:id/finalize", RoleMiddleware("student"), app.TeamHandler.FinalizeTeam)
				teams.POST("/:id/assign-advisor", RoleMiddleware("student"), app.TeamHandler.AssignAdvisor)
				teams.POST("/:id/advisor-response", RoleMiddleware("advisor"), app.TeamHandler.AdvisorResponse)
			}

			// Proposals (Students & Teachers)
			proposals := protected.Group("/proposals")
			{
				// Create a new Draft (Student Only)
				proposals.POST("", RoleMiddleware("student"), app.ProposalHandler.CreateProposal)

				// Update Draft OR Create Revision (Student Only)
				proposals.PUT("/:id", RoleMiddleware("student"), app.ProposalHandler.UpdateProposal)

				// Submit Proposal (Student Only - Leader)
				proposals.POST("/:id/submit", RoleMiddleware("student"), app.ProposalHandler.SubmitProposal)

				// View Proposals
				proposals.GET("", app.ProposalHandler.GetProposals)

				// View Specific Proposal Details
				proposals.GET("/:id", app.ProposalHandler.GetProposal)

				// View Version History
				proposals.GET("/:id/versions", app.ProposalHandler.GetVersions)

				// Create New Version (Student Only)
				proposals.POST("/:id/versions", RoleMiddleware("student"), app.ProposalHandler.CreateVersion)

				// Delete Draft (Student Only)
				proposals.DELETE("/:id", RoleMiddleware("student"), app.ProposalHandler.DeleteProposal)

				// Start Review (Advisor Only)
				proposals.POST("/:id/start-review", RoleMiddleware("advisor"), app.ProposalHandler.StartReview)
			}
			// Feedback (Teachers)
			feedback := protected.Group("/feedback")
			feedback.Use(RoleMiddleware("advisor"))
			{
				feedback.GET("/pending", app.FeedbackHandler.GetPendingProposals)
				feedback.POST("", app.FeedbackHandler.CreateFeedback)
				feedback.GET("/:id", app.FeedbackHandler.GetFeedback)
				
			}
 		protected.GET("/proposals/:id/feedback", app.FeedbackHandler.GetProposalFeedback)
			// Admin User Management
			admin := protected.Group("/admin")
			admin.Use(RoleMiddleware("admin"))
			{
				// User Management
				admin.POST("/users/teacher", app.UserHandler.CreateTeacher)
				admin.POST("/users/student", app.UserHandler.CreateStudent)
				admin.GET("/users", app.UserHandler.GetUsers)
				admin.GET("/advisors", app.UserHandler.GetAdvisors)
				admin.GET("/users/:id", app.UserHandler.GetUser)
				admin.PATCH("/users/:id/status", app.UserHandler.UpdateUserStatus)
				admin.POST("/users/:id/assign-department", app.UserHandler.AssignDepartment)
				admin.DELETE("/users/:id", app.UserHandler.DeleteUser)
				admin.GET("/stats", app.UserHandler.GetDashboardStats) 
				admin.PATCH("/proposals/:id/assign", app.ProposalHandler.AssignAdvisor)
				// Audit Logs
				admin.GET("/audit-logs", app.AuditHandler.GetAuditLogs)
				admin.GET("/audit-logs/:id", app.AuditHandler.GetAuditLog)
			}

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", app.NotificationHandler.GetNotifications)
				notifications.GET("/unread-count", app.NotificationHandler.GetUnreadCount)
				notifications.POST("/:id/mark-read", app.NotificationHandler.MarkAsRead)
				notifications.POST("/mark-all-read", app.NotificationHandler.MarkAllAsRead)
			}

			// AI Service
			ai := protected.Group("/ai")
			{
				ai.GET("/health", app.AIHandler.HealthCheck)
				ai.POST("/analyze-proposal", app.AIHandler.AnalyzeProposal)
				ai.GET("/check-similarity", RoleMiddleware("advisor"), app.AIHandler.CheckSimilarity)
			}

			// Projects (Team creators can manage, all can view)
			projects := protected.Group("/projects")
			{
				projects.POST("", app.ProjectHandler.CreateProject)
				projects.GET("", app.ProjectHandler.GetProjects)
				projects.GET("/:id", app.ProjectHandler.GetProject)
				projects.PUT("/:id", app.ProjectHandler.UpdateProject)
				projects.POST("/:id/publish", app.ProjectHandler.PublishProject)
				projects.POST("/:id/share", app.ProjectHandler.IncrementShareCount)
				// Project Reviews
				projects.POST("/:id/reviews", app.ReviewHandler.CreateReview)
				projects.GET("/:id/reviews", app.ReviewHandler.GetProjectReviews)
			}

			// Documentation
// Note: Changed projectId to id to match common patterns, 
// or keep projectId if you prefer. 
docsGroup := protected.Group("/projects/:id/documentation") 
{
    docsGroup.GET("", app.DocumentationHandler.GetProjectDocs)
    docsGroup.POST("", RoleMiddleware("student"), app.DocumentationHandler.Submit)
}
// Individual Doc Actions (For deleting or reviewing)
docActions := protected.Group("/documentation")
{
    docActions.DELETE("/:id", RoleMiddleware("student"), app.DocumentationHandler.Delete)
    docActions.PATCH("/:id/review", RoleMiddleware("advisor"), app.DocumentationHandler.Review)
}

			// // Documentation review (Teachers only)
			// docReview := protected.Group("/documentation")
			// docReview.Use(RoleMiddleware("teacher", "admin"))
			// {
			// 	docReview.POST("/:id/review", func(c *gin.Context) {
			// 		response.JSON(c, http.StatusNotImplemented, "review document not implemented", nil)
			// 	})
			// }

	
		}
	}

	return r
}
