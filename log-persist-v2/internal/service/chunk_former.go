package service

import (
	"log"
	"log-persist-v2/internal/cache"
	"time"
)

const (
	MaxTempQueueSizeMB    = 5               // Maximum size of tempQueue in MB
	MaxChunkFormationTime = 15 * time.Second // Maximum duration since last chunk formation
	ChunkCheckInterval    = 3 * time.Second // Interval to check the tempQueue size
)

func StartChunkFormer(tQ *cache.TmpQueue, cc *cache.ChunkCache) {
	var lastChunkFormationTime time.Time

	for {
		time.Sleep(ChunkCheckInterval)
		log.Print("chunkformer logging")
		lastChunkFormationTime,_ = tQ.MakeAndSaveChunk(cc, lastChunkFormationTime)
	}
}


