package main

import (
	"log"
	"log-persist-v2/internal/api"
	"log-persist-v2/internal/cache"
	"log-persist-v2/internal/service"
	"log-persist-v2/internal/kafka"
	"log-persist-v2/internal/s3"
	"net/http"

)

func main() {
	tmpQueue := cache.NewTmpQueue()
	chunkCache := cache.NewChunkCache()
	uploader, err := s3.NewS3Uploader("localhost:9004", "root", "qwertyuiop", "log-monitor")
	if err != nil {
		log.Fatalf("Failed to initialize S3 uploader: %v", err)
	}
	chunkFlusher  := service.NewChunkFlusher(chunkCache, "", uploader)
	go chunkFlusher.Start()
	
	go service.StartChunkFormer(tmpQueue, chunkCache)

	go kafka.StartKafkaConsumer("partitioned-topic", "localhost:9092", tmpQueue)



	http.HandleFunc("/filter", api.NewFilterHandler(chunkCache, tmpQueue))

	// Start HTTP server
	log.Println("Server started on :3002")
	if err := http.ListenAndServe(":3002", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}