package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Commit the processed messages. The method also removes old messages and offsets.
func (r Receiver) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	r.mu.Lock()

	threshold := 48 * time.Hour
	removeOldMessages(r.uncommittedMessages, threshold)
	removeOldOffcets(r.offcets, threshold)
	uncomMsg, ok := r.uncommittedMessages[msgUuid]
	if !ok {
		r.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	r.mu.Unlock()

	lastOffcetWithTimeStamp, ok := r.offcets[uncomMsg.msg.Partition]
	if !ok || lastOffcetWithTimeStamp.value < uncomMsg.msg.Offset {
		err := r.reader.CommitMessages(ctx, uncomMsg.msg)
		if err != nil {
			return fmt.Errorf("can not commit a message in the broker: %w", err)

		}
	}

	r.mu.Lock()
	r.offcets[uncomMsg.msg.Partition] = offcetWithTimeStamp{value: uncomMsg.msg.Offset, timeStamp: time.Now()}
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

func removeOldOffcets(offcets map[int]offcetWithTimeStamp, threashold time.Duration) {
	now := time.Now()

	for part, offcet := range offcets {
		if offcet.timeStamp.Add(threashold).Before(now) {
			delete(offcets, part)
		}
	}
}
