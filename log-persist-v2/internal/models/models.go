package models

import (
	"time"
	"github.com/google/uuid"
	"regexp"
)

// Event represents an event that will be ingested from Kafka
type Event struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

// Chunk represents a group of events that have been processed into a chunk
type Chunk struct {
	FirstLogTimestamp string `json:"first_log_timestamp"`
	LastLogTimestamp  string `json:"last_log_timestamp"`
	ChunkID           string `json:"chunk_id"`
	Data              string `json:"data"`
	Size  			  int64  `json:"size"`
}

// FilterRequest represents the request payload for filtering events via the API
type FilterRequest struct {
	Regex     string `json:"regex"`
	Contains  string `json:"contains"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// ChunkMetadata represents metadata related to the chunk cache
type ChunkMetadata struct {
	ChunkID           string `json:"chunk_id"`
	FirstLogTimestamp string `json:"first_log_timestamp"`
	LastLogTimestamp  string `json:"last_log_timestamp"`
}

// EventFilter represents the structure used for event filtering
type EventFilter struct {
	Regex    *regexp.Regexp
	Contains string
	StartTime time.Time
	EndTime   time.Time
}

// Helper function to generate a new chunk ID
func GenerateChunkID() string {
	return uuid.New().String()
}

// Helper function to parse timestamp from an event
func ParseTimestamp(timestamp string) (time.Time, error) {
	return time.Parse(time.RFC3339, timestamp)
}

// Helper function to extract the hour from a timestamp
func GetHourFromTimestamp(timestamp string) (int, error) {
	t, err := ParseTimestamp(timestamp)
	if err != nil {
		return 0, err
	}
	return t.Hour(), nil
}

// Helper function to determine if a chunk should be uploaded based on its size or age
func ShouldUploadChunk(chunkSize int64, lastUpload time.Time) bool {
	currentTime := time.Now()
	if chunkSize >= 100*1024*1024 || currentTime.Sub(lastUpload) >= 10*time.Minute {
		return true
	}
	return false
}
