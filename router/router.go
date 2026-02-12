package router

import (
	"net/http"
	"os"

	"lem-be/database"

	"github.com/gin-gonic/gin"
)

// Setup initializes and returns the Gin router with all routes configured
func Setup() *gin.Engine {
	// Set Gin mode from environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})

	// Initialize services and handlers
	loginService := auth_services.NewLoginService(database.GetDB())
	loginHandler := auth.NewLoginHandler(loginService)

	googleService := auth_services.NewGoogleService(*database.GetDB())
	googleHandler := auth.NewGoogleHandler(googleService)

	passwordResetService := auth_services.NewPasswordResetService(*database.GetDB())
	passwordResetHandler := auth.NewPasswordResetHandler(passwordResetService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Add your routes here
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		v1.POST("/login", loginHandler.HandleLogin)

		// OAuth2 routes
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/google/login", googleHandler.HandleGoogleLogin)
			authGroup.GET("/google/callback", googleHandler.HandleGoogleCallback)

			// Password reset routes
			authGroup.POST("/forgot-password", passwordResetHandler.HandleForgotPassword)
			authGroup.POST("/verify-otp", passwordResetHandler.HandleVerifyOTP)
			authGroup.POST("/reset-password", passwordResetHandler.HandleResetPassword)
		}
	}

	return router
}
