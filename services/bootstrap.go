package services

import (
	"context"
	"os"
	"time"

	"lem-be/constants"
	"lem-be/models"
	"lem-be/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InitSuperuser checks for an existing super_admin and creates one if it doesn't exist
func InitSuperuser(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := db.Collection("users")

	log := utils.NewLogger("BootstrapService", "InitSuperuser")
	// Check if any super_admin exists
	var existingSuperAdmin models.User
	err := usersCollection.FindOne(ctx, bson.M{"role": constants.RoleSuperAdmin}).Decode(&existingSuperAdmin)
	if err == nil {
		log.Info("Superuser already exists")
		return nil
	}

	if err != mongo.ErrNoDocuments {
		return err
	}

	// No super_admin found, create one from environment variables
	email := os.Getenv("SUPERUSER_EMAIL")
	password := os.Getenv("SUPERUSER_PASSWORD")

	if email == "" || password == "" {
		log.Warn("SUPERUSER_EMAIL or SUPERUSER_PASSWORD not set. Skipping superuser bootstrap.")
		return nil
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	superuser := models.User{
		Email:     email,
		Password:  hashedPassword,
		Role:      constants.RoleSuperAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = usersCollection.InsertOne(ctx, superuser)
	if err != nil {
		return err
	}

	log.Infof("Successfully bootstrapped superuser with email %s", email)
	return nil
}
