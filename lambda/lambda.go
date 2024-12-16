package main

import (
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"log"
	"regexp"
	"sync"
	"time"

	// "github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/service/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type LogEvent struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

type FilterRequest struct {
	S3Keys    []string `json:"s3Keys"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
	Regex     string   `json:"regex"`
	Contains  string   `json:"contains"`
}

type EventFilter struct {
	mu      sync.Mutex
	events  []LogEvent
}

// Global session to reuse across Lambda invocations
// var sess *session.Session
var client *minio.Client

func init() {
	// Initialize session once, this will be reused for subsequent invocations
	// ctx := context.Background()
	log.Printf("initting client")
	var err error
	client, err = minio.New("localhost:9004", &minio.Options{
		Creds:  credentials.NewStaticV4("root", "qwertyuiop", ""),Region: "us-east-1", Secure: false,
	
	})
	log.Printf("innited client")
	if err != nil {
		fmt.Errorf("failed to load S3 configuration: %w", err)
		log.Fatal("ailed to load S3 configuration:")
	}
}

// Add adds an event to the thread-safe array
func (ef *EventFilter) Add(event LogEvent) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.events = append(ef.events, event)
}

// filterEvent filters events based on the provided criteria
func filterEvent(event LogEvent, startTime, endTime, regex, contains string) bool {
	// Check if timestamp is within the given range
	log.Printf("start time %s",startTime)
	timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		log.Printf("Invalid timestamp: %v", event.Timestamp)
		return false
	}

	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		log.Printf("Invalid start time: %v", startTime)
		return false
	}

	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		log.Printf("Invalid end time: %v", endTime)
		return false
	}

	if timestamp.Before(start) || timestamp.After(end) {
		return false
	}

	// Check if the event data matches the regex
	if regex != "" {
		re, err := regexp.Compile(regex)
		if err != nil {
			log.Printf("Invalid regex: %v", regex)
			return false
		}
		if !re.MatchString(event.Data) {
			return false
		}
	}

	// Check if the event data contains the specified substring
	if contains != "" && !containsSubstring(event.Data, contains) {
		return false
	}

	return true
}

// containsSubstring checks if the event data contains the substring
func containsSubstring(data, contains string) bool {
	return strings.Contains(data, contains)
}

// stringContains checks if the string contains the substring
func stringContains(data, substr string) bool {
	return string(data) == substr
}

// downloadS3File downloads a file from MinIO (S3 emulator)
func downloadS3File(bucket, key string) ([]byte, error) {
	// Use the global session



	reader, err := client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})

	if err != nil {
		log.Printf("Error downloading file from s3. error : %s", err.Error())
	}

	downloadInfoBytes, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading file bytes from s3. error : %s", err.Error())
		}
	return downloadInfoBytes, nil


}

// Lambda handler function
func handler(ctx context.Context, request FilterRequest) ([]LogEvent, error) {
	// Create a thread-safe array for storing filtered events
	// var filter FilterRequest

	log.Printf("keys %s, regex  %s, contains %s, starts %s,ends : %s", request.S3Keys, request.Regex, request.Contains, request.StartTime, request.EndTime)


	eventFilter := &EventFilter{}

	var wg sync.WaitGroup

	// Download files concurrently and filter events
	for _, key := range request.S3Keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			log.Printf("downloading file %s",key)

			// Download the S3 file from MinIO
			fileContent, err := downloadS3File("log-monitor", key)
			if err != nil {
				log.Printf("Failed to download S3 file %s: %v", key, err)
				return
			}

			// Parse the file content as JSON array of events
			var events []LogEvent
			if err := json.Unmarshal(fileContent, &events); err != nil {
				log.Printf("Failed to parse JSON file %s: %v", key, err)
				return
			}

			// Filter the events
			for _, event := range events {
				if filterEvent(event, request.StartTime, request.EndTime, request.Regex, request.Contains) {
					eventFilter.Add(event)
				}
			}
		}(key)
	}

	// Wait for all downloads and processing to finish
	wg.Wait()

	// Return the filtered events
	return eventFilter.events, nil
}

func main() {
	// Start the Lambda function
	
	awslambda.Start(handler)
}
