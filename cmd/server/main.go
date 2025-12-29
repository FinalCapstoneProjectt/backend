package main

import (
	"backend/config"
	"backend/internal/app"
	"log"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	// 2. Bootstrap App (DB, Migrations, etc.)
	application, err := app.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("Bootstrap failed: %v", err)
	}

	// 3. Setup Router
	r := app.NewRouter(application.DB)

	// 4. Start Server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
