package handlers

import (
	"net/http"

	"lem-be/models"
	"lem-be/services"
	"lem-be/utils"

	"github.com/gin-gonic/gin"
)

type Login interface {
	HandleLogin(c *gin.Context)
}

type LoginHandler struct {
	loginService services.LoginService
}

func NewLoginHandler(loginService services.LoginService) Login {
	return &LoginHandler{loginService: loginService}
}


// HandleLogin is the Gin HTTP handler that adapts the business logic to Gin's handler format
func (h LoginHandler) HandleLogin(c *gin.Context) {
	var req models.LoginRequest
	log := utils.NewLogger("LoginHandler", "HandleLogin").WithContext(c.Request.Context())
	
	// Bind and validate the JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	log.Infof("Login attempt for email %s", req.Email)

	resp, err := h.loginService.Login(c.Request.Context(), req)
	if err != nil {
		// Handle authentication errors with 401 Unauthorized
		if err.Error() == "user not found" || err.Error() == "invalid password" {
			log.Warnf("Authentication failed for email %s: %s", req.Email, err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}
		
		// Handle other errors with 500 Internal Server Error
		log.Errorf("Login failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Login failed",
			"details": err.Error(),
		})
		return
	}

	log.Infof("Login successful for email %s", req.Email)

	c.JSON(http.StatusOK, resp)
}