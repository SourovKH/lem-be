package auth_services

import (
	"context"
	"errors"
	"time"

	auth_models "lem-be/auth/models"
	models "lem-be/auth/models"
	auth_utils "lem-be/auth/utils"

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

	// Find user by email
	var user auth_models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.LoginResponse{}, ErrUserNotFound
		}
		return models.LoginResponse{}, err
	}

	// Verify password
	if !auth_utils.ComparePasswords(user.Password, req.Password) {
		return models.LoginResponse{}, ErrInvalidPassword
	}

	// Generate access token
	accessToken, err := auth_utils.GenerateAccessToken(user.ID.Hex(), user.Email)
	if err != nil {
		return models.LoginResponse{}, ErrTokenGeneration
	}

	// Generate refresh token
	refreshToken, err := auth_utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
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