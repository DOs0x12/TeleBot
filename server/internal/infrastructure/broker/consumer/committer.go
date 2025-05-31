package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/DOs0x12/TeleBot/server/v3/retry"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func (kr KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	sObj, ok := kr.uncommittedMessageService.GetObj(msgUuid)
	if !ok {
		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}
	uncomMsg := sObj.Obj.(kafka.Message)
	err := kr.commitMesWithRetries(ctx, uncomMsg)
	if err != nil {
		return fmt.Errorf("failed to commit a message in the broker: %w", err)
	}

	kr.offsetService.AddOrUpdateOffset(uncomMsg.Partition, uncomMsg.Offset)
	kr.uncommittedMessageService.DelObj(msgUuid)

	return nil
}

func (kr KafkaConsumer) commitMesWithRetries(ctx context.Context, msg kafka.Message) error {
	act := func(ctx context.Context) error {
		lastOffsetWithTimeStamp, ok := kr.offsetService.GetOffset(msg.Partition)
		if ok && lastOffsetWithTimeStamp.Value > msg.Offset {
			return nil
		}

		return kr.reader.CommitMessages(ctx, msg)
	}
	rCnt := 5
	rDur := 1 * time.Second
	return retry.ExecuteWithRetries(ctx, act, rCnt, rDur)
}
