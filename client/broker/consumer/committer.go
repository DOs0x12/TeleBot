package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Commit commits the processed messages. The method also removes old messages and offsets.
func (r KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	r.mu.Lock()

	threshold := 48 * time.Hour
	removeOldMessages(r.uncommittedMessages, threshold)
	removeOldOffsets(r.offsets, threshold)
	uncomMsg, ok := r.uncommittedMessages[msgUuid]
	if !ok {
		r.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	r.mu.Unlock()

	lastOffsetWithTimeStamp, ok := r.offsets[uncomMsg.msg.Partition]
	if !ok || lastOffsetWithTimeStamp.value < uncomMsg.msg.Offset {
		err := r.reader.CommitMessages(ctx, uncomMsg.msg)
		if err != nil {
			return fmt.Errorf("failed to commit a message in the broker: %w", err)

		}
	}

	r.mu.Lock()
	r.offsets[uncomMsg.msg.Partition] = offsetWithTimeStamp{value: uncomMsg.msg.Offset, timeStamp: time.Now()}
	delete(r.uncommittedMessages, msgUuid)
	r.mu.Unlock()

	return nil
}

func removeOldMessages(uncommittedMessages map[uuid.UUID]uncommittedMessage, threshold time.Duration) {
	now := time.Now()

	for msgUuid, procMsg := range uncommittedMessages {
		if procMsg.timeStamp.Add(threshold).Before(now) {
			delete(uncommittedMessages, msgUuid)
		}
	}
}

func removeOldOffsets(offsets map[int]offsetWithTimeStamp, threashold time.Duration) {
	now := time.Now()

	for part, offset := range offsets {
		if offset.timeStamp.Add(threashold).Before(now) {
			delete(offsets, part)
		}
	}
}
