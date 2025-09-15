package kafka

import (
	"context"
	"encoding/json"
	"log"
	"log-persist-v2/internal/cache"
	"log-persist-v2/internal/models"
	// "time"

	"github.com/segmentio/kafka-go"
)

func StartKafkaConsumer(topic string, broker string, tmpQueue *cache.TmpQueue) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "consumer-group",
		StartOffset: kafka.LastOffset,
	})
	log.Printf("Starting kafka consumer")

	for {
		log.Printf("inside reader loop")
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}
		log.Printf("Kafka Message : %s",string(msg.Value))
		var event models.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshalling message from kafka. Dropping this event: %v", err)
			continue
		}
		tmpQueue.Enqueue(event)
		log.Printf("consumer %s", tmpQueue.ReadQueue())
	}
}
