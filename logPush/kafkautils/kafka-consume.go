// kafka/kafka.go
package kafkautils

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaClient struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string) (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	return producer, nil
}

// NewKafkaConsumer creates a new Kafka consumer
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

// ProduceMessage sends a message to a specific topic and partition
func (kc *KafkaClient) ProduceMessage(topic string, partition int32, message string) {
	if err := kc.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
		Key:            []byte(fmt.Sprintf("Key-%d", time.Now().Unix())),
		Value:          []byte(message),
	}, nil); err != nil {
		log.Printf("Failed to produce message: %v\n", err)
	}

	// Wait for delivery report
	e := <-kc.Producer.Events()
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		log.Printf("Delivery failed: %v\n", m.String())
	} else {
		log.Printf("\nProduced message to topic %s: key=%s value=%s\n", *m.TopicPartition.Topic, string(m.Key), string(m.Value))
	}
}

// ConsumeMessages reads messages from a Kafka topic
func (kc *KafkaClient) ConsumeMessages(topic string) {
	err := kc.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v\n", err)
	}

	for {
		msg, err := kc.Consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error: %v\n", err)
			break
		}
		log.Printf("Consumed message: key=%s value=%s from topic %s\n", string(msg.Key), string(msg.Value), *msg.TopicPartition.Topic)
	}
}
