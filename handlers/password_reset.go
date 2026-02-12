package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"lem-be/models"
	"lem-be/services"
	"lem-be/utils"

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
	log := utils.NewLogger("PasswordResetHandler", "HandleForgotPassword")

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnf("Invalid email address: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}

	log.Infof("Forgot password request for email %s", req.Email)

	err := h.passwordResetService.ForgotPassword(c, req)
	if err != nil {
		log.Errorf("Failed to send OTP to %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP", "details": err.Error()})
		return
	}

	log.Infof("OTP sent successfully to %s", req.Email)

	c.JSON(http.StatusOK, gin.H{"message": "If an account exists, an OTP has been sent."})
}

func (h *passwordResetHandler) HandleVerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	log := utils.NewLogger("PasswordResetHandler", "HandleVerifyOTP")

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Infof("Verifying OTP for email %s", req.Email)

	token, err := h.passwordResetService.VerifyOTP(c, req)
	if err != nil {
		log.Warnf("OTP verification failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP", "details": err.Error()})
		return
	}

	log.Infof("OTP verified successfully for email %s", req.Email)

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully", "reset_token": token})
}

// HandleResetPassword updates the user's password in the database
func (h *passwordResetHandler) HandleResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	log := utils.NewLogger("PasswordResetHandler", "HandleResetPassword")

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Info("Resetting password using token")

	if err := h.passwordResetService.ResetPassword(c, req); err != nil {
		log.Errorf("Failed to reset password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password", "details": err.Error()})
		return
	}

	log.Info("Password reset successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
