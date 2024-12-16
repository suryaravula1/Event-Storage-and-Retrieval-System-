// main.go
package main

import (
	// "fmt"
	"log"
	"time"
	"logPush/kafkautils"
	"encoding/json"

	// "os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
)

type Event struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

func main() {
	// Step 1: Create Kafka Admin Client
	// kafka_host := os.Getenv("KAFKA_HOST")
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092", // Replace with the first broker address
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka admin client: %s\n", err)
	}
	defer adminClient.Close()

	// Step 2: Initialize Kafka with partitions and get the config from init.go
	kafkaConfig := kafkautils.InitKafka(adminClient)

	// Step 3: Create Kafka producer and consumer clients
	producer, err := kafkautils.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %s\n", err)
	}
	defer producer.Close()

	consumer, err := kafkautils.NewKafkaConsumer(kafkaConfig.Brokers, "go-consumer-group")
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %s\n", err)
	}
	defer consumer.Close()

	// Initialize Kafka client struct
	kafkaClient := kafkautils.KafkaClient{
		Producer: producer,
		Consumer: consumer,
	}

	// Step 4: Set up Fiber server to receive logs from Fluentd
	app := fiber.New()

	// Define a route for log ingestion
	app.Post("/api", func(c *fiber.Ctx) error {
		// Get the raw JSON body from the request
		jsonMessage := c.Body()
		log.Printf("this is body XXXXXXXX %s", jsonMessage)

		// Parse the raw JSON body into a slice of Event structs
		var events []Event
		if err := json.Unmarshal(jsonMessage, &events); err != nil {
			return c.Status(400).SendString("Invalid JSON format" + err.Error())
		}

		// Process each event
		for i := range events {
			// If timestamp is missing or empty, set it to the current UTC timestamp
			if events[i].Timestamp == "" {
				events[i].Timestamp = time.Now().UTC().Format(time.RFC3339)
			}

			// Marshal the event into JSON string
			updatedEventJSON, err := json.Marshal(events[i])
			if err != nil {
				return c.Status(500).SendString("Error marshaling event")
			}

			// P

		// Produce the raw JSON to Kafka
		kafkaClient.ProduceMessage(kafkaConfig.Topic, 0, string(updatedEventJSON))
	}

		return c.SendString("Log received and sent to Kafka")
	})

	// Start Fiber server to listen for logs from Fluentd
	go func() {
		log.Fatal(app.Listen("localhost:3003"))
	}()
	select {}

	// Step 5: Consume messages from Kafka topic
	// kafkaClient.ConsumeMessages(kafkaConfig.Topic)
}
