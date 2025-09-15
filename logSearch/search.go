package main
import(
	"log"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"fmt"
	"sync"
	"encoding/json"
)


var (
	eventArray []Event
	arrayMutex sync.Mutex
)
type Event struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}
type KafkaConfig struct {
	Topic             string
	NumPartitions     int
	ReplicationFactor int
	Brokers           []string
}
func start(){

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092", // Replace with the first broker address
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka admin client: %s\n", err)
	}
	defer adminClient.Close()

	// Step 2: Initialize Kafka with partitions and get the config from init.go
		kafkaConfig := KafkaConfig{
		Topic:             "partitioned-topic",
		NumPartitions:     2,                             // Adjust the partition count as needed
		ReplicationFactor: 1,                             // Adjust replication factor as per requirements
		Brokers:           []string{"kafka:9092"}, // Replace with the first broker address
	}

	consumer, err := NewKafkaConsumer(kafkaConfig.Brokers, "go-consumer-group")
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %s\n", err)
	}
	defer consumer.Close()

	// Initialize Kafka client struct
	kafkaClient := KafkaClient{
		Consumer: consumer,
	}

	kafkaClient.ConsumeMessages(kafkaConfig.Topic)
}
type KafkaClient struct {
	Consumer *kafka.Consumer
}
func NewKafkaConsumer(brokers []string, groupID string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers[0],
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}
	return consumer, nil
}
func (kc *KafkaClient) ConsumeMessages(topic string) {
	err := kc.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v\n", err)
	}

	for {
		msg, err := kc.Consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error while reading: %v\n", err)
			break
		}

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Safely add event to the global array
		arrayMutex.Lock()
		eventArray = append(eventArray, event)
		arrayMutex.Unlock()
		log.Printf("Consumed message: key=%s value=%s from topic %s\n", string(msg.Key), string(msg.Value), *msg.TopicPartition.Topic)
	}
}

func ReadEvents() []Event {
	arrayMutex.Lock()
	defer arrayMutex.Unlock()

	// Copy the events and clear the array
	events := make([]Event, len(eventArray))
	copy(events, eventArray)
	eventArray = []Event{}

	return events
}