package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/DOs0x12/TeleBot/server/internal/common/retry"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func (kr KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	kr.mu.Lock()

	threshold := 48 * time.Hour
	removeOldMessages(kr.uncommittedMessages, threshold)
	removeOldOffsets(kr.offsets, threshold)
	uncomMsg, ok := kr.uncommittedMessages[msgUuid]
	if !ok {
		kr.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	kr.mu.Unlock()

	err := kr.commitMesWithRetries(ctx, uncomMsg.msg)
	if err != nil {
		return fmt.Errorf("can not commit a message in the broker: %w", err)

	}

	kr.mu.Lock()
	kr.offsets[uncomMsg.msg.Partition] = offsetWithTimeStamp{value: uncomMsg.msg.Offset, timeStamp: time.Now()}
	delete(kr.uncommittedMessages, msgUuid)
	kr.mu.Unlock()

	return nil
}

func (kr *KafkaConsumer) commitMesWithRetries(ctx context.Context, msg kafka.Message) error {
	act := func(ctx context.Context) error {
		lastOffsetWithTimeStamp, ok := kr.offsets[msg.Partition]
		if ok && lastOffsetWithTimeStamp.value > msg.Offset {
			return nil
		}

		return kr.reader.CommitMessages(ctx, msg)
	}

	return retry.ExecuteWithRetries(ctx, act)
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
