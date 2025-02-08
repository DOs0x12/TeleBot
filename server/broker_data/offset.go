package broker_data

import (
	"context"
	"sync"
	"time"
)

// OffsetService stores partition offsets of a broker.
type OffsetService struct {
	mu      *sync.Mutex
	offsets map[int]offsetWithTimeStamp
}

type offsetWithTimeStamp struct {
	Value     int64
	timeStamp time.Time
}

// NewOffsetService creates a new OffsetService instance.
func NewOffsetService() OffsetService {
	return OffsetService{mu: &sync.Mutex{}, offsets: make(map[int]offsetWithTimeStamp)}
}

// StartCleanupOffsets starts a worker for cleaning up the old offsets.
func (s OffsetService) StartCleanupOffsets(ctx context.Context) {
	go s.delOldOffsetsWithPeriod(ctx)
}

func (s OffsetService) delOldOffsetsWithPeriod(ctx context.Context) {
	t := time.NewTicker(1 * time.Hour)
	threshold := 48 * time.Hour
	for {
		select {
		case <-ctx.Done():
			t.Stop()

			return
		case <-t.C:
			s.removeOldOffsets(threshold)
		}
	}
}

// AddOrUpdateOffset adds or updates an offset to others.
func (s OffsetService) AddOrUpdateOffset(partition int, offset int64) {
	s.mu.Lock()
	s.offsets[partition] = offsetWithTimeStamp{Value: offset, timeStamp: time.Now()}
	s.mu.Unlock()
}

// GetOffset gets an offset by partition.
func (s OffsetService) GetOffset(partition int) (offsetWithTimeStamp, bool) {
	s.mu.Lock()
	offset, ok := s.offsets[partition]
	s.mu.Unlock()
	return offset, ok
}

func (s OffsetService) removeOldOffsets(threashold time.Duration) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for part, offset := range s.offsets {
		if offset.timeStamp.Add(threashold).Before(now) {
			delete(s.offsets, part)
		}
	}
}
