package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client

func init() {
	// Initialize MongoDB connection
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://microservices-demo-birthdays-mongo:27017/birthdaydb"
	}
	clientOptions := options.Client().ApplyURI(mongoURL)
	var err error
	client, err = mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	// Initialize RabbitMQ connection
	initRabbitMQ()
}

func main() {
	r := gin.Default()

	// Setup API endpoints
	r.GET("/birthdays", GetBirthdays(client))
	r.GET("/birthdays/:id", GetBirthdayByID(client))
	r.POST("/birthdays", CreateBirthday(client))
	r.PUT("/birthdays/:id", UpdateBirthday(client))
	r.DELETE("/birthdays/:id", DeleteBirthday(client))

	// Start RabbitMQ listener in a separate goroutine
	go StartListeningForEvents(client) // Changed this to the correct function call

	// Start the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
