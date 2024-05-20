package main

import (
	"context"
	"log"
	"net/http"
	"reflect"
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

		if birthday.ID == "" {
			birthday.ID = uuid.New().String()
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
		id := c.Param("id")

		// Convert the string ID to a UUID
		uuidID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var birthday Birthday
		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = collection.FindOne(ctx, bson.M{"id": uuidID.String()}).Decode(&birthday) // Use the string representation of the UUID
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
        id := c.Param("id")
        uuidID, err := uuid.Parse(id)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
            return
        }

        var reqBirthday Birthday
        if err := c.ShouldBindJSON(&reqBirthday); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        collection := getCollection(client)
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        update := bson.M{"$set": bson.M{"birthday": reqBirthday.Birthday}}
        result, err := collection.UpdateOne(ctx, bson.M{"id": uuidID.String()}, update, options.Update().SetUpsert(true))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        if result.MatchedCount == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "Birthday not found"})
            return
        }

        // Optionally log the updated document for debugging
        updatedDoc := Birthday{}
        if err := collection.FindOne(ctx, bson.M{"id": uuidID.String()}).Decode(&updatedDoc); err != nil {
            log.Printf("Error finding updated birthday: %v", err)
        }

        c.JSON(http.StatusOK, updatedDoc)
    }
}


func DeleteBirthday(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		uuidID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		collection := getCollection(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = collection.DeleteOne(ctx, bson.M{"id": uuidID.String()})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
