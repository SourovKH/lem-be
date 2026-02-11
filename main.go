package main

import (
	"lem-be/database"
	"lem-be/router"
	"log"
	"os"
)

func main() {
	// Initialize MongoDB connection
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize Gin router
	r := router.Setup()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
