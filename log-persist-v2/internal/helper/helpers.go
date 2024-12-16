package helper

import (
	"encoding/json"
	"log"
	"log-persist-v2/internal/models"
	"strconv"
	"time"
)

// groupEventsByHour groups events based on their "hour" value derived from the timestamp.
func GroupEventsByHour(events []models.Event) map[int][]models.Event {
	grouped := make(map[int][]models.Event)

	for _, event := range events {
		hour, err := models.GetHourFromTimestamp(event.Timestamp)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}

		grouped[hour] = append(grouped[hour], event)
	}

	return grouped
}

// processGroupedEvents processes a group of events, returning the first timestamp,
// last timestamp, and concatenated event data.
func ProcessGroupedEvents(events []models.Event) (string, string, string) {
	if len(events) == 0 {
		return "", "", ""
	}

	// Extract the first and last timestamps
	firstTimestamp := events[0].Timestamp
	lastTimestamp := events[len(events)-1].Timestamp

	// Marshal events into JSON array
	jsonData, err := json.Marshal(events)
	if err != nil {
		log.Printf("Error marshaling events to JSON: %v", err)
		return "", "", ""
	}

	// Return timestamps and JSON array of events
	return firstTimestamp, lastTimestamp, string(jsonData)
}

func GenerateS3Key(chunk models.Chunk, hour int) string {
	// Parse the string timestamp into time.Time
	firstLogTime, err := time.Parse(time.RFC3339, chunk.FirstLogTimestamp)
	if err != nil {
		log.Printf("Error parsing FirstLogTimestamp: %v", err)
		return ""
	}

	// Format the parsed time into "YYYY-MM-DD"
	formattedDate := firstLogTime.UTC().Format("2006-01-02")

	// Construct the S3 key
	return formattedDate + "-" + strconv.Itoa(hour) + "-" + chunk.ChunkID
}