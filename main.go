package main

import (
	"context"
	"os"

	"lem-be/database"
	"lem-be/router"
	"lem-be/services"
	"lem-be/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		// We don't have the logger yet, so we use standard log here or just ignore if it's fine
	}

	// Initialize Global Logger
	utils.InitLogger("AuthServer", "Main")
	log := utils.NewLogger("AuthServer", "Main")

	// Initialize Tracer
	shutdown, err := utils.InitTracer()
	if err != nil {
		log.Errorf("Failed to initialize tracer: %v", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				log.Errorf("Error shutting down tracer: %v", err)
			}
		}()
	}

	// Initialize MongoDB connection
	if err := database.Init(); err != nil {
		log.Errorf("Failed to initialize MongoDB: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Bootstrap superuser
	log.Info("Bootstrapping superuser...")	
	if err := services.InitSuperuser(database.GetDB()); err != nil {
		log.Errorf("Error bootstrapping superuser: %v", err)
	}
	log.Info("Superuser bootstrapped successfully")

	// Initialize OAuth2 config
	utils.InitOAuthConfig()

	// Initialize Gin router
	r := router.Setup()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Infof("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Errorf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
