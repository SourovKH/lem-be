package database

import (
	"context"
	"lem-be/utils"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Init initializes the MongoDB connection
func Init() error {
	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	utils.NewLogger("Database", "Init").Info("Successfully connected to MongoDB")

	Client = client
	
	// Get database name from environment or use default
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "lem_db"
	}
	Database = client.Database(dbName)

	return nil
}

// Close disconnects from MongoDB
func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := Client.Disconnect(ctx); err != nil {
			return err
		}
		utils.NewLogger("Database", "Close").Info("Successfully disconnected from MongoDB")
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *mongo.Database {
	return Database
}

// GetClient returns the MongoDB client instance
func GetClient() *mongo.Client {
	return Client
}
