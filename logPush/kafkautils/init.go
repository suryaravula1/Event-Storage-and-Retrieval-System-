package kafkautils

import (
	"context"
	"fmt"
	"log"


	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaConfig holds the configuration for Kafka initialization
type KafkaConfig struct {
	Topic             string
	NumPartitions     int
	ReplicationFactor int
	Brokers           []string
}

// InitKafka initializes the Kafka topic with partitions and replication
func InitKafka(adminClient *kafka.AdminClient) KafkaConfig {
	// Define topic and partition settings
	kafkaConfig := KafkaConfig{
		Topic:             "partitioned-topic",
		NumPartitions:     2,                             // Adjust the partition count as needed
		ReplicationFactor: 1,                             // Adjust replication factor as per requirements
		Brokers:           []string{"localhost:9092"}, // Replace with the first broker address
	}

	// Use Metadata to check if the topic exists
	topicName := kafkaConfig.Topic
	metadata, err := adminClient.GetMetadata(&topicName, false, 5000)
	if err != nil {
		log.Printf("Error occurred while getting Kafka metadata: %s\n", err)
		return kafkaConfig
	}

	// Check if the topic exists in the metadata
	topicExists := false
	if topicMetadata, ok := metadata.Topics[topicName]; ok {
		topicExists = true
		if len(topicMetadata.Partitions) == 0 {
			log.Printf("Topic '%s' has zero partitions, recreating...\n", topicName)
			topicExists = false
		} else {
			fmt.Printf("Topic '%s' already exists with %d partitions.\n", topicName, len(topicMetadata.Partitions))
		}
	}

	// Create the topic if it does not exist or has zero partitions
	if !topicExists {
		topicSpecification := kafka.TopicSpecification{
			Topic:             kafkaConfig.Topic,
			NumPartitions:     kafkaConfig.NumPartitions,
			ReplicationFactor: kafkaConfig.ReplicationFactor,
		}

		_, err := adminClient.CreateTopics(
			context.Background(),
			[]kafka.TopicSpecification{topicSpecification},
		)
		if err != nil {
			log.Fatalf("Failed to create topic: %s\n", err)
		}
		fmt.Printf("Topic '%s' created with %d partitions.\n", kafkaConfig.Topic, kafkaConfig.NumPartitions)
	}

	return kafkaConfig
}
