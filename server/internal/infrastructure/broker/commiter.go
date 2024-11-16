package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func (kr KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	kr.mu.Lock()

	kr.removeOldMessages()
	procMsg, ok := kr.processingMessages[msgUuid]
	if !ok {
		kr.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	kr.mu.Unlock()

	err := kr.commitMesWithRetries(ctx, procMsg.msg)
	if err != nil {
		return fmt.Errorf("can not commit a message in the broker: %w", err)

	}

	kr.mu.Lock()
	delete(kr.processingMessages, msgUuid)
	kr.mu.Unlock()

	return nil
}

func (kr KafkaConsumer) commitMesWithRetries(ctx context.Context, msg kafka.Message) error {
	act := func(ctx context.Context) error {
		return kr.reader.CommitMessages(ctx, msg)
	}

	return common.ExecuteWithRetries(ctx, act)
}

func (kr KafkaConsumer) removeOldMessages() {
	now := time.Now()

	for msgUuid, procMsg := range kr.processingMessages {
		if procMsg.timeStamp.Add(48 * time.Hour).Before(now) {
			delete(kr.processingMessages, msgUuid)
		}
	}
}
