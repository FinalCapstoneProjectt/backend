package app

import (
	"backend/config"
	"backend/internal/domain"
	"backend/pkg/database"
	"log"

	"gorm.io/gorm"
)

type App struct {
	Config config.Config
	DB     *gorm.DB
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
	)
	if err != nil {
		return nil, err
	}
	log.Println("Database migration completed")

	// 3. Initialize Services (DI)
	// Example: authService := auth.NewService(auth.NewRepository(db))

	return &App{
		Config: cfg,
		DB:     db,
	}, nil
}
