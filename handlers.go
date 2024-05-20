package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("birthdaydb").Collection("birthdays")
}

func CreateBirthday(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var birthday Birthday
		if err := c.ShouldBindJSON(&birthday); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if birthday.ID == uuid.Nil {
			birthday.ID = uuid.New()
		}

		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, birthday)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, birthday)
	}
}

func GetBirthdayByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var birthday Birthday
		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = collection.FindOne(ctx, bson.M{"id": id}).Decode(&birthday)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Birthday not found"})
			return
		}

		c.JSON(http.StatusOK, birthday)
	}
}

func GetBirthdays(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var birthdays []Birthday
		collection := getCollection(client)
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
}

func UpdateBirthday(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var birthday Birthday
		if err := c.ShouldBindJSON(&birthday); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		update := bson.M{"$set": birthday}
		_, err = collection.UpdateOne(ctx, bson.M{"id": id}, update, options.Update().SetUpsert(true))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, birthday)
	}
}

func DeleteBirthday(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = collection.DeleteOne(ctx, bson.M{"id": id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
