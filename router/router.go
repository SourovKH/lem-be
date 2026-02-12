package router

import (
	"net/http"
	"os"

	"lem-be/database"
	"lem-be/handlers"
	"lem-be/services"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Setup initializes and returns the Gin router with all routes configured
func Setup() *gin.Engine {
	// Set Gin mode from environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add OpenTelemetry middleware FIRST before any routes
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "auth-server"
	}
	router.Use(otelgin.Middleware(serviceName))
	router.Use(TraceLogger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})

	// Initialize services and handlers
	loginService := services.NewLoginService(database.GetDB())
	loginHandler := handlers.NewLoginHandler(loginService)

	googleService := services.NewGoogleService(*database.GetDB())
	googleHandler := handlers.NewGoogleHandler(googleService)

	passwordResetService := services.NewPasswordResetService(*database.GetDB())
	passwordResetHandler := handlers.NewPasswordResetHandler(passwordResetService)

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
