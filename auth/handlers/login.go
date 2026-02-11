package auth

import (
	"net/http"

	models "lem-be/auth/models"
	auth_services "lem-be/auth/services"

	"github.com/gin-gonic/gin"
)

type Login interface {
	HandleLogin(c *gin.Context)
}

type LoginHandler struct {
	loginService auth_services.LoginService
}

func NewLoginHandler(loginService auth_services.LoginService) Login {
	return &LoginHandler{loginService: loginService}
}


// HandleLogin is the Gin HTTP handler that adapts the business logic to Gin's handler format
func (h LoginHandler) HandleLogin(c *gin.Context) {
	var req models.LoginRequest
	
	// Bind and validate the JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.loginService.Login(c.Request.Context(), req)
	if err != nil {
		// Handle authentication errors with 401 Unauthorized
		if err.Error() == "user not found" || err.Error() == "invalid password" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}
		
		// Handle other errors with 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Login failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}