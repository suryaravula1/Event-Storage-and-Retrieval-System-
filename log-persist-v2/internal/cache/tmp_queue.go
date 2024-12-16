package cache

import (
	"log"
	"sync"
	"time"

	// "time"
	"log-persist-v2/internal/helper"
	"log-persist-v2/internal/models"
)
const (
	MaxTempQueueSizeMB    = 5               // Maximum size of tempQueue in MB
	MaxChunkFormationTime = 15 * time.Second // Maximum duration since last chunk formation
	ChunkCheckInterval    = 3 * time.Second // Interval to check the tempQueue size
)
type TmpQueue struct {
	mu    sync.Mutex
	queue []models.Event
	size  int64 // in bytes
}

func NewTmpQueue() *TmpQueue {
	return &TmpQueue{queue: []models.Event{}}
}

func (tq *TmpQueue) Enqueue(event models.Event) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tq.queue = append(tq.queue, event)
	tq.size += int64(len(event.Data))
}

func (tq *TmpQueue) ReadQueue() []models.Event {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.queue

}

// func (tq *TmpQueue) DequeueAll() ([]models.Event, int64) {
// 	tq.Mu.Lock()
// 	defer tq.Mu.Unlock()

// 	data := tq.Queue
// 	size := tq.size
// 	tq.Queue = []models.Event{}
// 	tq.size = 0
// 	return data, size
// }

func (tq *TmpQueue) Size() int64 {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.size
}

func (tq *TmpQueue) SizeMB() float64 {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	// Convert size in bytes to megabytes
	return float64(tq.size) / (1024 * 1024)
}

func (tq *TmpQueue) MakeAndSaveChunk(cc *ChunkCache, lastChunkFormationTime time.Time)(time.Time, error){
	
	tq.mu.Lock()
	defer tq.mu.Unlock()

	//dequeue
	shouldFormChunk := tq.size >= MaxTempQueueSizeMB || time.Since(lastChunkFormationTime) >= MaxChunkFormationTime
		log.Printf("Should form chunk ?  %v",shouldFormChunk)
		if shouldFormChunk {
	
	if len(tq.queue) > 0 {
		log.Print("Adding chunk to Chunk Cache")
		eventGroups:=helper.GroupEventsByHour(tq.queue)
		log.Printf("grpd events : %v", eventGroups)
		for hour, groupedEvents := range eventGroups {
			firstTimestamp, lastTimestamp, concatenatedData := helper.ProcessGroupedEvents(groupedEvents)

			chunk := models.Chunk{
				FirstLogTimestamp: firstTimestamp,
				LastLogTimestamp:  lastTimestamp,
				ChunkID:           models.GenerateChunkID(),
				Data:              concatenatedData,
			}
			log.Printf("Fromed chunk: %v", chunk)
			cc.AddChunk(hour, chunk)

			//update qu size
			tq.size = 0
			tq.queue = []models.Event{}
		}
		return time.Now(),nil
	}else {
		log.Printf("no event to form chunk from")
		return lastChunkFormationTime,nil
	}
}
return lastChunkFormationTime, nil
}


