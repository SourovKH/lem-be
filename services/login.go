package services

import (
	"context"
	"errors"
	"time"

	"lem-be/models"
	"lem-be/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrTokenGeneration  = errors.New("failed to generate tokens")
)

type LoginService interface {
	Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error)
}

type LoginServiceImpl struct {
	db *mongo.Database
}

func NewLoginService(db *mongo.Database) LoginService {
	return &LoginServiceImpl{db: db}
}

func (s *LoginServiceImpl) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	// Get users collection
	usersCollection := s.db.Collection("users")

	log := utils.NewLogger("LoginService", "Login")
	// Find user by email
	var user models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warnf("User not found for email %s", req.Email)
			return models.LoginResponse{}, ErrUserNotFound
		}
		log.Errorf("Database error during user lookup for email %s: %v", req.Email, err)
		return models.LoginResponse{}, err
	}

	// Verify password
	if !utils.ComparePasswords(user.Password, req.Password) {
		log.Warnf("Invalid password attempt for email %s", req.Email)
		return models.LoginResponse{}, ErrInvalidPassword
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		log.Errorf("Failed to generate access token for email %s: %v", req.Email, err)
		return models.LoginResponse{}, ErrTokenGeneration
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		log.Errorf("Failed to generate refresh token for email %s: %v", req.Email, err)
		return models.LoginResponse{}, ErrTokenGeneration
	}

	// Update user's last login time (optional)
	_, _ = usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)

	return models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}