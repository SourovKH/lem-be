package handlers

import (
	"net/http"

	"lem-be/services"
	"lem-be/utils"

	"github.com/gin-gonic/gin"
)

type GoogleHandler interface {
	HandleGoogleLogin(c *gin.Context)
	HandleGoogleCallback(c *gin.Context)
}

type googleHandler struct {
	googleService services.GoogleService
}

func NewGoogleHandler(googleService services.GoogleService) GoogleHandler {
	return &googleHandler{googleService: googleService}
}

// HandleGoogleLogin redirects the user to Google's OAuth2 login page
func (h *googleHandler) HandleGoogleLogin(c *gin.Context) {
	log := utils.NewLogger("GoogleHandler", "HandleGoogleLogin").WithContext(c.Request.Context())
	// In production, use a secure random state and store it in session/cookie to prevent CSRF
	state := "random-state"
	url := utils.GoogleOAuthConfig.AuthCodeURL(state)
	log.Info("Redirecting to Google Login")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGoogleCallback handles the callback from Google, fetches user info, and issues a JWT
func (h *googleHandler) HandleGoogleCallback(c *gin.Context) {
	log := utils.NewLogger("GoogleHandler", "HandleGoogleCallback").WithContext(c.Request.Context())
	user, accessToken, refreshToken, err := h.googleService.HandleGoogleCallback(c)
	if err != nil {
		log.Errorf("Failed to handle Google callback: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle Google callback", "details": err.Error()})
		return
	}

	log.Infof("Google login successful for email %s", user.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"email": user.Email,
			"role":  user.Role,
		},
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
