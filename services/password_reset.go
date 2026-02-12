package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"lem-be/models"
	"lem-be/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
)

type PasswordResetService interface {
	ForgotPassword(c *gin.Context, req models.ForgotPasswordRequest) error
	VerifyOTP(c *gin.Context, req models.VerifyOTPRequest) (string, error)
	ResetPassword(c *gin.Context, req models.ResetPasswordRequest) error
}

type passwordResetService struct {
	db mongo.Database
}

func NewPasswordResetService(db mongo.Database) PasswordResetService {
	return &passwordResetService{db: db}
}

// HandleForgotPassword generates an OTP and sends it via email
func (h *passwordResetService) ForgotPassword(c *gin.Context, req models.ForgotPasswordRequest) error {
	ctx, span := otel.Tracer("password-reset-service").Start(c.Request.Context(), "ForgotPassword")
	defer span.End()

	log := utils.NewLogger("PasswordResetService", "ForgotPassword").WithContext(ctx)
	// 1. Verify user exists and is a local user
	usersCollection := h.db.Collection("users")
	var user models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		log.Warnf("User not found for password reset (email: %s)", req.Email)
		// Security: Don't reveal if email exists or not, just return 200
		return errors.New("If an account exists, an OTP has been sent.")
	}

	if user.Provider != "" && user.Provider != "local" {
		log.Warnf("Attempted password reset for non-local user (email: %s, provider: %s)", req.Email, user.Provider)
		return errors.New("This account uses social login. Please use the social provider to sign in.")
	}

	// 2. Generate 6-digit OTP
	otp, _ := generateOTP()

	// 3. Save OTP to database
	otpCollection := h.db.Collection("otps")
	otpRecord := models.OTPRecord{
		Email:     req.Email,
		Code:      otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	_, err = otpCollection.UpdateOne(
		context.Background(),
		bson.M{"email": req.Email},
		bson.M{"$set": otpRecord},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		log.Errorf("Failed to store OTP in database for email %s: %v", req.Email, err)
		return errors.New("Failed to store OTP")
	}
	log.Infof("Successfully stored OTP for email %s", req.Email)

	// 4. Send Email
	if err := utils.SendOTPEmail(req.Email, otp); err != nil {
		log.Errorf("Failed to send OTP email to %s: %v", req.Email, err)
		// Don't fail the request, just log it. In dev, we can see the code in logs.
	} else {
		log.Infof("OTP email sent successfully to %s", req.Email)
	}

	return nil
}

// HandleVerifyOTP checks if the code is valid and issues a reset token
func (h *passwordResetService) VerifyOTP(c *gin.Context, req models.VerifyOTPRequest) (string, error) {
	ctx, span := otel.Tracer("password-reset-service").Start(c.Request.Context(), "VerifyOTP")
	defer span.End()

	var otpRecord models.OTPRecord
	log := utils.NewLogger("PasswordResetService", "VerifyOTP").WithContext(ctx)
	err := h.db.Collection("otps").FindOne(context.Background(), bson.M{
		"email": req.Email,
		"code":  req.Code,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&otpRecord)

	if err != nil {
		log.Warnf("Invalid or expired OTP attempt for email %s", req.Email)
		return "", errors.New("Invalid or expired OTP")
	}

	// Success! Delete the OTP so it can't be reused
	h.db.Collection("otps").DeleteOne(context.Background(), bson.M{"email": req.Email})

	// Issue a temporary Reset Token (using the same JWT utility but with short expiry)
	// We'll reuse GenerateAccessToken but maybe add a specific "reset" claim in a real app
	// For now, let's just generate a standard token that identifies the user
	token, err := utils.GenerateAccessToken("RESET:"+req.Email, req.Email, "reset_only")
	if err != nil {
		log.Errorf("Failed to generate reset token for email %s: %v", req.Email, err)
		return "", errors.New("Failed to generate reset token")
	}

	return token, nil
}

// HandleResetPassword updates the user's password in the database
func (h *passwordResetService) ResetPassword(c *gin.Context, req models.ResetPasswordRequest) error {
	ctx, span := otel.Tracer("password-reset-service").Start(c.Request.Context(), "ResetPassword")
	defer span.End()

	// 1. Verify Reset Token
	log := utils.NewLogger("PasswordResetService", "ResetPassword").WithContext(ctx)
	claims, err := utils.ValidateToken(req.ResetToken)
	if err != nil || claims.Role != "reset_only" {
		log.Warnf("Invalid or expired reset token: %s", req.ResetToken)
		return errors.New("Invalid or expired reset token")
	}

	// 2. Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("Failed to hash password")
	}

	// 3. Update user in MongoDB
	_, err = h.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"email": claims.Email},
		bson.M{"$set": bson.M{"password": hashedPassword, "updated_at": time.Now()}},
	)

	if err != nil {
		log.Errorf("Failed to update password in database for email %s: %v", claims.Email, err)
		return errors.New("Failed to update password")
	}
	log.Infof("Password updated successfully in database for email %s", claims.Email)

	return nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
