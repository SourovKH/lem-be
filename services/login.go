package services

import (
	"context"
	"errors"
	"time"

	auth_models "lem-be/models"
	auth_utils "lem-be/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrTokenGeneration  = errors.New("failed to generate tokens")
)

type LoginService interface {
	Login(ctx context.Context, req auth_models.LoginRequest) (auth_models.LoginResponse, error)
}

type LoginServiceImpl struct {
	db *mongo.Database
}

func NewLoginService(db *mongo.Database) LoginService {
	return &LoginServiceImpl{db: db}
}

func (s *LoginServiceImpl) Login(ctx context.Context, req auth_models.LoginRequest) (auth_models.LoginResponse, error) {
	// Get users collection
	usersCollection := s.db.Collection("users")

	// Find user by email
	var user auth_models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return auth_models.LoginResponse{}, ErrUserNotFound
		}
		return auth_models.LoginResponse{}, err
	}

	// Verify password
	if !auth_utils.ComparePasswords(user.Password, req.Password) {
		return auth_models.LoginResponse{}, ErrInvalidPassword
	}

	// Generate access token
	accessToken, err := auth_utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		return auth_models.LoginResponse{}, ErrTokenGeneration
	}

	// Generate refresh token
	refreshToken, err := auth_utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return auth_models.LoginResponse{}, ErrTokenGeneration
	}

	// Update user's last login time (optional)
	_, _ = usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)

	return auth_models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}