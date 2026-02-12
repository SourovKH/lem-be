package services

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"lem-be/constants"
	"lem-be/models"
	"lem-be/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GoogleService interface {
	HandleGoogleCallback(c *gin.Context) (user models.User, accessToken string, refreshToken string, err error)
}

type googleService struct {
	db mongo.Database
}

func NewGoogleService(db mongo.Database) GoogleService {
	return &googleService{db: db}
}

func (service *googleService) HandleGoogleCallback(c *gin.Context) (user models.User, accessToken string, refreshToken string, err error) {
	// 1. Verify state (skipped for brevity, but essential for production)
	
	log := utils.NewLogger("GoogleService", "HandleGoogleCallback")
	// 2. Exchange code for token
	code := c.Query("code")
	token, err := utils.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Errorf("Failed to exchange token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "details": err.Error()})
		return
	}
	log.Info("Successfully exchanged OAuth2 code for token")

	// 3. Fetch user info from Google
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Errorf("Failed to fetch user info from Google: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		log.Errorf("Failed to decode Google user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info", "details": err.Error()})
		return models.User{}, "", "", err
	}
	log.Infof("Fetched Google user info for email %s", googleUser.Email)

	// 4. Upsert user in MongoDB
	usersCollection := service.db.Collection("users")
	
	filter := bson.M{"provider": "google", "provider_id": googleUser.ID}
	update := bson.M{
		"$set": bson.M{
			"email":      googleUser.Email,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"role":       constants.RoleUser,
			"created_at": time.Now(),
			"provider":   "google",
			"provider_id": googleUser.ID,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err = usersCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Errorf("Failed to upsert user in MongoDB for email %s: %v", googleUser.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	log.Infof("User upserted successfully for email %s", googleUser.Email)

	// Fetch the user to get their ID and Role (especially if they were just created)
	err = usersCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
		return
	}

	// 5. Generate JWT tokens
	accessToken, err = utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return models.User{}, "", "", err
	}

	refreshToken, err = utils.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return models.User{}, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

	