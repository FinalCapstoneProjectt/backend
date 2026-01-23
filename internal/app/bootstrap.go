package app

import (
	"backend/config"
	"backend/internal/auth"
	"backend/internal/departments"

	//"backend/internal/documentation"
	"backend/internal/domain"
	//"backend/internal/feedback" // Commented out until ready
	"backend/internal/projects"
	"backend/internal/proposals"
	"backend/internal/teams"
	"backend/internal/universities"
	"backend/internal/users"
	"backend/pkg/audit"
	"backend/pkg/database"
	"log"

	"gorm.io/gorm"
)

type App struct {
	Config            config.Config
	DB                *gorm.DB
	AuditLogger       *audit.Logger
	AuthService       auth.Service
	AuthHandler       *auth.Handler
	UniversityHandler *universities.Handler
	DepartmentHandler *departments.Handler
	UserHandler       *users.Handler
	TeamHandler       *teams.Handler
	ProposalHandler   *proposals.Handler
	// FeedbackHandler   *feedback.Handler // Commented out to match initialization
	ProjectHandler    *projects.Handler
	// DocumentationHandler *documentation.Handler
}

func Bootstrap(cfg config.Config) (*App, error) {
	// 1. Connect to Database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	// 2. Automigrate Models
	err = db.AutoMigrate(
		&domain.University{},
		&domain.Department{},
		&domain.User{},
		&domain.Team{},
		&domain.TeamMember{},
		&domain.Proposal{},
		&domain.ProposalVersion{},
		&domain.Feedback{},
		&domain.Project{},
		&domain.ProjectDocumentation{},
		&domain.ProjectReview{},
		&domain.Notification{},
		&domain.AuditLog{},
	)
	if err != nil {
		return nil, err
	}
	log.Println("Database migration completed")

	// 3. Seed Database with Initial Data
	log.Println("Starting database seeding...")
	if err := database.SeedDatabase(db); err != nil {
		log.Printf("ERROR: Failed to seed database: %v", err)
	} else {
		log.Println("Database seeding completed successfully")
	}

	// 4. Initialize Audit Logger
	auditLogger := audit.NewLogger(db)
	log.Println("Audit logger initialized")

	// 4. Initialize Services (DI)
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg, auditLogger)
	authHandler := auth.NewHandler(authService)
	log.Println("Authentication service initialized")

	// 5. Initialize University Service
	universityRepo := universities.NewRepository(db)
	universityService := universities.NewService(universityRepo)
	universityHandler := universities.NewHandler(universityService)
	log.Println("University service initialized")

	// 6. Initialize Department Service
	departmentRepo := departments.NewRepository(db)
	departmentService := departments.NewService(departmentRepo)
	departmentHandler := departments.NewHandler(departmentService)
	log.Println("Department service initialized")

	// 7. Initialize User Service
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)
	log.Println("User service initialized")

	// 8. Initialize Team Service
	teamRepo := teams.NewRepository(db)
	teamService := teams.NewService(teamRepo)
	teamHandler := teams.NewHandler(teamService)
	log.Println("Team service initialized")

	// 9. Initialize Proposal Service
	proposalRepo := proposals.NewRepository(db)
	// ⚠️ FIXED: Added 'db' argument for transaction support
	proposalService := proposals.NewService(proposalRepo, db) 
	proposalHandler := proposals.NewHandler(proposalService)
	log.Println("Proposal service initialized")

	// 10. Initialize Feedback Service (Commented out per your snippet)
	// feedbackRepo := feedback.NewRepository(db)
	// feedbackService := feedback.NewService(feedbackRepo, proposalRepo)
	// feedbackHandler := feedback.NewHandler(feedbackService)
	// log.Println("Feedback service initialized")

	// 11. Initialize Project Service
	projectRepo := projects.NewRepository(db)
	// Ensure Project Service signature matches. Assuming it takes proposalRepo.
	// If Project Service also needs DB now, check internal/projects/service.go
	projectService := projects.NewService(projectRepo, proposalRepo) 
	projectHandler := projects.NewHandler(projectService)
	log.Println("Project service initialized")

	// 12. Initialize Documentation Service
	//documentationRepo := documentation.NewRepository(db)
	//documentationService := documentation.NewService(documentationRepo, projectRepo)
	//documentationHandler := documentation.NewHandler(documentationService)
	//log.Println("Documentation service initialized")

	return &App{
		Config:            cfg,
		DB:                db,
		AuditLogger:       auditLogger,
		AuthService:       authService,
		AuthHandler:       authHandler,
		UniversityHandler: universityHandler,
		DepartmentHandler: departmentHandler,
		UserHandler:       userHandler,
		TeamHandler:       teamHandler,
		ProposalHandler:   proposalHandler,
		// FeedbackHandler:   feedbackHandler, // Must be commented out if not initialized above
		ProjectHandler:    projectHandler,
		// DocumentationHandler: documentationHandler,
	}, nil
}