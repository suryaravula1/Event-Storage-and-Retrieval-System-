package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	// "sync"
	"time"

	"log-persist-v2/internal/cache"
	"log-persist-v2/internal/models"
)

type FilterRequest struct {
	Regex     string `json:"regex"`
	Contains  string `json:"contains"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func NewFilterHandler(cc *cache.ChunkCache, tmpQueue *cache.TmpQueue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var req FilterRequest
		queryParams := r.URL.Query()

		// Extract specific query parameters
		reg := queryParams.Get("regex")
		contains := queryParams.Get("contains")
		startTime_ := queryParams.Get("startTime")
		endTime_ := queryParams.Get("endTime")

		// Handle missing parameters
		if startTime_ == "" || endTime_ == "" {
			http.Error(w, "Missing required query parameters", http.StatusBadRequest)
			return
		}

		// Compile regex if provided
		var regex *regexp.Regexp
		if reg != "" {
			var err error
			regex, err = regexp.Compile(reg)
			if err != nil {
				http.Error(w, "Invalid regex", http.StatusBadRequest)
				return
			}
		}

		// Parse start and end times
		startTime, err := time.Parse(time.RFC3339, startTime_)
		if err != nil {
			http.Error(w, "Invalid startTime format", http.StatusBadRequest)
			return
		}
		endTime, err := time.Parse(time.RFC3339, endTime_)
		if err != nil {
			http.Error(w, "Invalid endTime format", http.StatusBadRequest)
			return
		}

		// Thread-safe access to chunkCache and tmpQueue
		req :=  FilterRequest{
			StartTime: startTime_,
			EndTime: endTime_,
			Regex: reg,
			Contains: contains,}
		var filteredEvents []models.Event
		var processedChunkIDs []string
		// var mu sync.Mutex

		// wg := &sync.WaitGroup{}

		// Process chunkCache
		// wg.Add(1)
		func() {
			// defer wg.Done()
			log.Printf("chunk cache size%f", cc.SizeMB())

			for _, chunks := range cc.ReadCache() {
				for _, chunk := range chunks {
					// Extract individual events from chunk.Data
					var events []models.Event
					if err := json.Unmarshal([]byte(chunk.Data), &events); err != nil {
						log.Printf("Unable to unmarshal json data : %s", chunk.Data)
						continue // Skip if data cannot be unmarshaled
					}

					for _, event := range events {
						eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
						if err != nil {
							log.Printf("Error parsing Timestamp. Skipping this event: %v", err)
							continue
						}
						
						if eventTime.After(startTime) && eventTime.Before(endTime) {
							if matchesFilter(event.Data, regex, req.Contains) {
								// mu.Lock()
								filteredEvents = append(filteredEvents, event)
								// mu.Unlock()
							}
						}
					}
					// mu.Lock()
					processedChunkIDs = append(processedChunkIDs, chunk.ChunkID)
					// mu.Unlock()
				}
			}
		}()

		// Process tmpQueue
		// wg.Add(1)
		go func() {
			// defer wg.Done()
			for _, event := range tmpQueue.ReadQueue(){
				eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
				if err != nil {
					log.Printf("Error parsing Timestamp. Skipping this event: %v", err)
					continue
				}

				if eventTime.After(startTime) && eventTime.Before(endTime) {
					if matchesFilter(event.Data, regex, req.Contains) {
						// mu.Lock()
						filteredEvents = append(filteredEvents, event)
						// mu.Unlock()
					}
				}
			}
		}()

		// wg.Wait()

		// Prepare the response
		resp := map[string]interface{}{
			"events":    filteredEvents,
			"chunk_ids": processedChunkIDs,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// matchesFilter checks if an event matches the regex and contains filter
func matchesFilter(eventData string, regex *regexp.Regexp, contains string) bool {
	if regex != nil && !regex.MatchString(eventData) {
		return false
	}
	if contains != "" && !strings.Contains(eventData, contains) {
		return false
	}
	return true
}

