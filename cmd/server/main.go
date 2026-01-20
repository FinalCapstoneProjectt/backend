package main

// @title University Project Hub API
// @version 1.0
// @description REST API for managing universities, departments, teams, proposals, and reviews.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"backend/config"
	"backend/docs"
	"backend/internal/app"
	"log"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	// 1.1 Configure Swagger metadata at runtime
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.BasePath = "/api/v1"

	// 2. Bootstrap App (DB, Migrations, Services, etc.)
	application, err := app.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("Bootstrap failed: %v", err)
	}

	// 3. Setup Router with full app context
	r := app.NewRouter(application)

	// 4. Start Server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
