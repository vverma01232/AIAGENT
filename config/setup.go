package config

import (
	"aiagent/repository"
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger = logrus.New().WithField("app", "aiagent")

// ConnectDB function is used to instantiate MongoDB Connection
func ConnectDB() {
	URI := os.Getenv("MONGOURI")
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))

	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Connected to MongoDB")
	DB = client
}

// DB Client instance
var DB *mongo.Client

// GetCollection function helps in getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("AIAGENT").Collection(collectionName)
	return collection
}
func GetRepoCollection(collectionName string) repository.Repository {
	repo := repository.MongoUserRepository{
		Collection: GetCollection(DB, collectionName),
	}
	return &repo
}
func GetPromptCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("AgenticAI").Collection("AIPrompts")
	return collection
}
