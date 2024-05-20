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
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://db:27017/birthdaydb"
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
}

func main() {
	r := gin.Default()

	r.GET("/birthdays", GetBirthdays(client))
	r.GET("/birthdays/:id", GetBirthdayByID(client))
	r.POST("/birthdays", CreateBirthday(client))
	r.PUT("/birthdays/:id", UpdateBirthday(client))
	r.DELETE("/birthdays/:id", DeleteBirthday(client))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
