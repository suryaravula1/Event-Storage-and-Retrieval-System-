package cache

import (
	// "fmt"
	"context"
	"log"
	"log-persist-v2/internal/models"
	"log-persist-v2/internal/helper"
	"log-persist-v2/internal/s3"
	"sync"
	"time"
	// "time"
	// "github.com/google/uuid"
)

// type Chunk struct {
// 	FirstLogTimestamp string
// 	LastLogTimestamp  string
// 	ChunkID           string
// 	Data              string
// }
const MaxCacheSize = 200*1024*1024
type ChunkCache struct {
	mu     sync.Mutex
	chunks map[int]map[string]models.Chunk
	SizeBytes   int64 // in bytes
}

func (cc *ChunkCache) ReadCache() map[int]map[string]models.Chunk{
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.chunks
}

func NewChunkCache() *ChunkCache {
	return &ChunkCache{chunks: make(map[int]map[string]models.Chunk)}
}

func (cc *ChunkCache) AddChunk(hour int, chunk models.Chunk) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	log.Printf("Adding chunk to cache size : %d", chunk.Size)
	// Check if the hour key exists in the map
	if _, exists := cc.chunks[hour]; !exists {
		// Initialize the entry with an array containing the single chunk
		cc.chunks[hour] = map[string]models.Chunk{chunk.ChunkID:chunk}
	} else {
		// Append the chunk to the existing array
		cc.chunks[hour][chunk.ChunkID] = chunk
	}

	// Update the cache size
	cc.SizeBytes += int64(len(chunk.Data))
	log.Printf("chunk cache size after adding chunk: %d Bytes", cc.SizeBytes)
}


// func (cc *ChunkCache) GetAndRemoveChunks() map[string]*models.Chunk {
// 	cc.mu.Lock()
// 	defer cc.mu.Unlock()

// 	chunks := cc.chunks
// 	cc.chunks = make(map[int]*models.Chunk)
// 	cc.size = 0
// 	return chunks
// }

// func (cc *ChunkCache) Size() int64 {
// 	cc.mu.Lock()
// 	defer cc.mu.Unlock()
// 	return cc.SizeBytes
// }

// func extractTimestamp(event string) string {
// 	// Placeholder to parse timestamp from the event
// 	return time.Now().Format(time.RFC3339)
// }

func (cc *ChunkCache) SizeMB() float64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	// Convert size in bytes to megabytes
	return float64(cc.SizeBytes) / (1024 * 1024)
}


func (cc *ChunkCache) FlushChunksIfNeeded(lastFlush time.Time, s3uploader s3.S3Uploader) time.Time {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	log.Printf("flushing chunks if needed...: chunk size in Bytes : %d", cc.SizeBytes)

	// Check conditions: size >= 200 MB or 5 minutes since the last flush
	timeSinceLastFlush := time.Since(lastFlush)

	if cc.SizeBytes < MaxCacheSize && timeSinceLastFlush < 5*time.Second {
		return lastFlush
	}

	log.Printf("starting chunk flush...: chunk size in Bytes : %d", cc.SizeBytes)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for hour, chunks := range cc.chunks {
		// Process chunks concurrently
		func(hour int, chunks map[string]models.Chunk) {
			chunks_to_remove :=  make([]string, 0)
			for chunk_id, chunk := range chunks {
				key := helper.GenerateS3Key(chunk, hour)
				err := s3uploader.UploadToS3(ctx, key, chunk)
				if err != nil {
					log.Printf("Failed to upload chunk %s: %v\n", chunk.ChunkID, err)
					continue
				}
				chunks_to_remove = append(chunks_to_remove,chunk_id)

				log.Printf("Successfully uploaded chunk %s to S3\n", chunk.ChunkID)
			}
			log.Printf("Removing successfully uploaded chunks from cache !!")
			for _,c := range chunks_to_remove{
				size := cc.chunks[hour][c].Size
				delete(cc.chunks[hour], c)
				cc.SizeBytes -= size

			}
		}(hour, chunks)
	}

	
	log.Println("Chunk flush completed.")
	return time.Now()
}