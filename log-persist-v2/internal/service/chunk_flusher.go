package service

import (
	"log"
	"time"

	// "github.com/google/uuid"
	"log-persist-v2/internal/cache"
	"log-persist-v2/internal/s3" // Hypothetical package for S3 interaction
)

type ChunkFlusher struct {
	chunkCache   *cache.ChunkCache
	lastFlush    time.Time
	s3BucketName string
	s3uploader   s3.S3Uploader
}

func NewChunkFlusher(chunkCache *cache.ChunkCache, s3BucketName string, s3uploader *s3.S3Uploader) *ChunkFlusher {
	return &ChunkFlusher{
		chunkCache:   chunkCache,
		lastFlush:    time.Now(),
		s3BucketName: s3BucketName,
		s3uploader:   *s3uploader,
	}
}

func (cf *ChunkFlusher) Start() {

	for {
		time.Sleep(5*time.Second)
		log.Print("chunk flush time...")
		cf.lastFlush = cf.chunkCache.FlushChunksIfNeeded(cf.lastFlush,cf.s3uploader)
	}
}
