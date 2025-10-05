package main

import (
	"log"
	"log-persist-v2/internal/api"
	"log-persist-v2/internal/cache"
	"log-persist-v2/internal/service"
	"log-persist-v2/internal/kafka"
	"log-persist-v2/internal/s3"
	"net/http"
	"os"
)

func main() {
	tmpQueue := cache.NewTmpQueue()
	chunkCache := cache.NewChunkCache()
	
	// Get environment variables with defaults
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = "localhost:9004"
	}
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	if s3AccessKey == "" {
		s3AccessKey = "root"
	}
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	if s3SecretKey == "" {
		s3SecretKey = "qwertyuiop"
	}
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "log-monitor"
	}
	
	uploader, err := s3.NewS3Uploader(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket)
	if err != nil {
		log.Fatalf("Failed to initialize S3 uploader: %v", err)
	}
	chunkFlusher  := service.NewChunkFlusher(chunkCache, "", uploader)
	go chunkFlusher.Start()
	
	go service.StartChunkFormer(tmpQueue, chunkCache)

	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		kafkaHost = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "partitioned-topic"
	}
	go kafka.StartKafkaConsumer(kafkaTopic, kafkaHost, tmpQueue)

	http.HandleFunc("/filter", api.NewFilterHandler(chunkCache, tmpQueue))

	// Start HTTP server
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8082"
	}
	log.Printf("Server started on :%s", serverPort)
	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}