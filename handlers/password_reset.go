package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"lem-be/models"
	"lem-be/services"

	"github.com/gin-gonic/gin"
)

type PasswordResetHandler interface {
	HandleForgotPassword(c *gin.Context)
	HandleVerifyOTP(c *gin.Context)
	HandleResetPassword(c *gin.Context)
}

type passwordResetHandler struct {
	passwordResetService services.PasswordResetService
}

func NewPasswordResetHandler(passwordResetService services.PasswordResetService) PasswordResetHandler {
	return &passwordResetHandler{passwordResetService: passwordResetService}
}

func (h *passwordResetHandler) HandleForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}

	err := h.passwordResetService.ForgotPassword(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If an account exists, an OTP has been sent."})
}

func (h *passwordResetHandler) HandleVerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, err := h.passwordResetService.VerifyOTP(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully", "reset_token": token})
}

// HandleResetPassword updates the user's password in the database
func (h *passwordResetHandler) HandleResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.passwordResetService.ResetPassword(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
