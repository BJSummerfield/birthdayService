package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func getCollection() *mongo.Collection {
    return client.Database("birthdaydb").Collection("birthdays")
}

func createBirthday(c *gin.Context) {
    var birthday Birthday
    if err := c.ShouldBindJSON(&birthday); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if birthday.ID == primitive.NilObjectID {
        birthday.ID = primitive.NewObjectID()
    }

    collection := getCollection()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, birthday)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, birthday)
}

func getBirthday(c *gin.Context) {
    id := c.Param("id")
    objID, _ := primitive.ObjectIDFromHex(id)

    var birthday Birthday
    collection := getCollection()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&birthday)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, birthday)
}

func getAllBirthdays(c *gin.Context) {
    var birthdays []Birthday
    collection := getCollection()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var birthday Birthday
        cursor.Decode(&birthday)
        birthdays = append(birthdays, birthday)
    }

    c.JSON(http.StatusOK, birthdays)
}

func updateBirthday(c *gin.Context) {
    id := c.Param("id")
    objID, _ := primitive.ObjectIDFromHex(id)

    var birthday Birthday
    if err := c.ShouldBindJSON(&birthday); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    collection := getCollection()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    update := bson.M{"$set": birthday}
    _, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update, options.Update().SetUpsert(true))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, birthday)
}

func deleteBirthday(c *gin.Context) {
    id := c.Param("id")
    objID, _ := primitive.ObjectIDFromHex(id)

    collection := getCollection()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
