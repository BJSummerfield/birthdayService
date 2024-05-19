package main

import (
    "context"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
    // Set up MongoDB connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://db:27017"))
    if err != nil {
        log.Fatal(err)
    }

    // Set up Gin router
    router := gin.Default()

    router.POST("/birthdays", createBirthday)
    router.GET("/birthdays/:id", getBirthday)
    router.GET("/birthdays", getAllBirthdays)
    router.PUT("/birthdays/:id", updateBirthday)
    router.DELETE("/birthdays/:id", deleteBirthday)

    router.Run(":8080")
}
