package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
        amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EventMessage struct for RabbitMQ events
type EventMessage struct {
	Timestamp     string      `json:"timestamp"`
	Version       string      `json:"version"`
	ServiceOrigin string      `json:"serviceOrigin"`
	TraceID       uuid.UUID   `json:"traceId"`
	EventType     string      `json:"eventType"`
	Environment   string      `json:"environment"`
	Payload       interface{} `json:"payload"`
}

// Setup RabbitMQ connection with retry logic
func setupRabbitMQ() *amqp.Channel {
	var conn *amqp.Connection
	var err error
	attempts := 0
	for attempts < 30 {
		conn, err = amqp.Dial("amqp://rabbitmq:5672/")
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, attempt %d: %v", attempts+1, err)
		time.Sleep(5 * time.Second)
		attempts++
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after 30 attempts: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		"user_events", // exchange name
		"topic",       // exchange type
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	log.Println("Connected to RabbitMQ and channel opened.")
	return ch
}

func PublishEvent(eventType string, payload interface{}) {
	ch := setupRabbitMQ()
	defer ch.Close()

	event := EventMessage{
		Timestamp:     time.Now().Format(time.RFC3339),
		Version:       "1.0",
		ServiceOrigin: "BirthdayService",
		TraceID:       uuid.New(),
		EventType:     eventType,
		Environment:   "production",
		Payload:       payload,
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = ch.Publish(
		"user_events",                 // exchange
		"userManagement."+eventType,   // routing key
		false,                         // mandatory
		false,                         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
		return
	}

	log.Printf("Published %s event: %s", eventType, string(body))
}

func StartListeningForEvents(client *mongo.Client) {
    ch := setupRabbitMQ()
    defer ch.Close()

    queue, err := ch.QueueDeclare(
        "birthdayQueue",  // Queue name can be left empty to let RabbitMQ generate a unique name
        true,             // Make the queue durable so it survives broker restarts
        false,            // Do not delete when unused
        false,            // Exclusive should be false to allow connections from other consumers in the future
        false,            // No-wait
        nil,              // Arguments
    )
    failOnError(err, "Failed to declare a queue")

    // Binding to userCreated events
    err = ch.QueueBind(
        queue.Name,               // queue name
        "userManagement.userCreated", // specific routing key for userCreated
        "user_events",            // exchange
        false,                    // no-wait
        nil,                      // arguments
    )
    failOnError(err, "Failed to bind a queue for userCreated")

    // Binding to userDeleted events
    err = ch.QueueBind(
        queue.Name,               // queue name
        "userManagement.userDeleted", // specific routing key for userDeleted
        "user_events",            // exchange
        false,                    // no-wait
        nil,                      // arguments
    )
    failOnError(err, "Failed to bind a queue for userDeleted")

    msgs, err := ch.Consume(
        queue.Name,   // queue
        "",           // consumer tag
        false,        // turn off auto-ack, consider manual ack after successful processing
        false,        // exclusive
        false,        // no-local
        false,        // no-wait
        nil,          // args
    )
    failOnError(err, "Failed to register a consumer")

    log.Println("Successfully connected to RabbitMQ and waiting for messages.")

    forever := make(chan bool)
    go func() {
        for d := range msgs {
            log.Printf("Received a message: %s", d.Body)
            handleMessage(client, d.Body)
            d.Ack(false) // Acknowledge message only after successful handling
        }
    }()
    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
}

func handleMessage(client *mongo.Client, body []byte) {
	var msg EventMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Error decoding message: %v", err)
		return
	}

	switch msg.EventType {
	case "userCreated":
		userID, ok := msg.Payload.(map[string]interface{})["id"].(string)
		if !ok {
			log.Printf("Invalid payload for userCreated event")
			return
		}
		createBirthdayRecord(client, userID)
	case "userDeleted":
		userID, ok := msg.Payload.(map[string]interface{})["id"].(string)
		if !ok {
			log.Printf("Invalid payload for userDeleted event")
			return
		}
		deleteBirthdayRecord(client, userID)
	default:
		log.Printf("Unhandled event type: %s", msg.EventType)
	}
}

func createBirthdayRecord(client *mongo.Client, userID string) {
	collection := client.Database("birthdaydb").Collection("birthdays")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc := bson.M{
		"id":       userID,
		"birthday": nil,
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		log.Printf("Failed to create birthday record for user %s: %v", userID, err)
		return
	}
	log.Printf("Birthday record created for user ID: %s", userID)
}

func deleteBirthdayRecord(client *mongo.Client, userID string) {
	collection := client.Database("birthdaydb").Collection("birthdays")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"id": userID})
	if err != nil {
		log.Printf("Failed to delete birthday record for user %s: %v", userID, err)
		return
	}
	log.Printf("Birthday record deleted for user ID: %s", userID)
}

