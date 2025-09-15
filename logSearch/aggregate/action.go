package aggregate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/net/context"
)

type Event struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}
var re = regexp.MustCompile(`\d{4}-\d{2}-\d{2}-\d{2}-([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`)

type PersistResp struct {
	Events []Event `json:"events"`
	ChunkIds []string `json:"chunk_ids"`
}
type LambdaReq struct {
	S3Keys    []string `json:"s3Keys"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
	Regex     string   `json:"regex"`
	Contains  string   `json:"contains"`
}

type FilterRequest struct {
	Regex     string `json:"regex"`
	Contains  string `json:"contains"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}


// Redis setup (change to your Redis configuration)
var Rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

var s3Client, _ = minio.New("localhost:9004", &minio.Options{
	Creds:  credentials.NewStaticV4("root", "qwertyuiop", ""),
	Secure: false,
})

// Channel size for handling responses
const channelSize = 10

func AccumulateResults(filter FilterRequest, uuid string) {

	log.Printf("Request = %v", filter)
	// call persist
	persitResp := callPersist(filter)
	filteredData, err := json.Marshal(persitResp.Events)
	if err != nil {
		log.Fatalf("Failed to marshal filtered logs: %v", err)
	}
	ctx := context.Background()
	err = Rdb.Set(ctx, uuid, filteredData, 0).Err()
	if err != nil {
		log.Printf("Failed to write logs to Redis: %v", err)
	} else {
		log.Printf("Logs written to Redis with UUID: %s", uuid)
	}

	// put in redis

	// save chunkids

	// list s3 objects 

		// Generate all possible prefixes between startPrefix and endPrefix
		var prefixes []string
		start,_:= time.Parse(time.RFC3339,filter.StartTime)
		end,_:= time.Parse(time.RFC3339, filter.EndTime)
		for t := start; !t.After(end); t = t.Add(time.Hour) {
			prefix := fmt.Sprintf("%d/%02d/%02d/%02d/", t.Year(), t.Month(), t.Day(), t.Hour())
			prefixes = append(prefixes, prefix)
		}
		var objKeys []string
		var s3chunkids []string
		for _, k := range prefixes{
			opts := minio.ListObjectsOptions{
				Recursive: true,
				Prefix:    k,
			}

			for object := range s3Client.ListObjects(context.Background(), "log-monitor", opts) {
				if object.Err != nil {
					fmt.Println(object.Err)
					return
				}
				


				// Find UUID in the string
				uuid := re.FindString(object.Key)
				if uuid == "" {
					fmt.Println("No UUID found in the input string")
				} else {
					fmt.Println("Extracted UUID:", uuid)
					if !contains(persitResp.ChunkIds,uuid) {
						s3chunkids = append(s3chunkids, uuid)
						objKeys = append(objKeys,object.Key)
					}
					
				}
			}
		}
	


	// form chunkids from object keys



	// skip objvets with processed chinkids

	// divide chunks into lambdas
		callLambdas(objKeys,filter, uuid)
	


		
	// make lambda parallel req

	// keep adding to redis



}


func callPersist(req FilterRequest) PersistResp{ 
	url := fmt.Sprintf(
		"http://localhost:3002/query?regex=%s&contains=%s&startTime=%s&endTime=%s",
		req.Regex,
		req.Contains,
		req.StartTime,
		req.EndTime,
	)// Replace with your actual URL

	// // Marshal the request data
	// // jsonData, err := json.Marshal(req)
	// if err != nil {
	// 	log.Println("Error marshaling request data:", err)
	// 	return
	// }
	var persistResp PersistResp
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error making HTTP request:", err)
		return persistResp
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	
	err = json.Unmarshal(body,&persistResp)
	if err != nil {
		log.Println("Error marshalling resp from persist HTTP request:", err)
		return persistResp

	}

	return persistResp


}


// Function to divide a slice of strings into n equal parts
func divideIntoParts(array []string, n int) [][]string {
	length := len(array)
	partSize := length / n
	remainder := length % n

	result := make([][]string, 0, n)
	start := 0

	for i := 0; i < n; i++ {
		extra := 0
		if i < remainder {
			extra = 1 // Distribute remaining elements among the first parts
		}

		end := start + partSize + extra
		result = append(result, array[start:end])
		start = end
	}

	return result
}


func callLambdas(objKeys []string, filter FilterRequest, uuid string ) {
	parts := divideIntoParts(objKeys, 5)
	ctx := context.Background()

	// // Example HTTP request data
	// requests := []FilterRequest{
	// 	{Regex: ".*error.*", Contains: "critical", StartTime: "2024-11-18T12:00:00Z", EndTime: "2024-11-18T13:00:00Z"},
	// 	{Regex: ".*debug.*", Contains: "info", StartTime: "2024-11-18T12:30:00Z", EndTime: "2024-11-18T13:30:00Z"},
	// }

	// Create a channel to store responses
	responseChannel := make(chan string, channelSize)
	done := make(chan bool)

	// Start a goroutine to periodically append to Redis
	go func() {
		cache := []string{}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case data := <-responseChannel:
				cache = append(cache, data)
				if len(cache) >= 5 { // Threshold for bulk appending to Redis
					appendToRedis(ctx, cache, uuid)
					cache = nil
				}
			case <-done: // Final append when all requests are done
				if len(cache) > 0 {
					appendToRedis(ctx, cache, uuid)
				}
				return
			case <-ticker.C: // Periodic check for remaining data
				if len(cache) > 0 {
					appendToRedis(ctx, cache, uuid)
					cache = nil
				}
			}
		}
	}()

	// Make HTTP requests in goroutines
	for _, key := range parts {
		req := LambdaReq{
			S3Keys: key,
			StartTime: filter.StartTime,
			EndTime: filter.EndTime,
			Regex: filter.Regex,
			Contains: filter.Contains,
		}
		go makeHTTPRequest(req, responseChannel)
	}

	// Wait for all responses (simulating with sleep here)
	time.Sleep(5 * time.Second)
	close(done)
}



func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}


func makeHTTPRequest(req LambdaReq, responseChannel chan<- string) {
	url := "http://localhost:8080/2015-03-31/functions/function/invocations" // Replace with your actual URL

	// Marshal the request data
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Println("Error marshaling request data:", err)
		return
	}

	// Make the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	// Send the response to the channel
	responseChannel <- string(body)
}


func appendToRedis(ctx context.Context, data []string, uuid string) {
	if len(data) == 0 {
		return
	}

	// Append data to Redis (list)
	_, err := Rdb.RPush(ctx, "responses", data).Result()
	if err != nil {
		log.Println("Error appending to Redis:", err)
	} else {
		fmt.Println("Appended to Redis:", data)
	}
}