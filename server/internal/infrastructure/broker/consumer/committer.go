package consumer

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/internal/common/retry"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func (kr KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	uncomMsg, ok := kr.getMsgFromUncommited(msgUuid)
	if !ok {
		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	err := kr.commitMesWithRetries(ctx, uncomMsg.msg)
	if err != nil {
		return fmt.Errorf("failed to commit a message in the broker: %w", err)

	}

	kr.addOrUpdateOffset(uncomMsg.msg.Partition, uncomMsg.msg.Offset)
	kr.delMsgFromUncommitted(msgUuid)

	return nil
}

func (kr *KafkaConsumer) commitMesWithRetries(ctx context.Context, msg kafka.Message) error {
	act := func(ctx context.Context) error {
		lastOffsetWithTimeStamp, ok := kr.getOffset(msg.Partition)
		if ok && lastOffsetWithTimeStamp.value > msg.Offset {
			return nil
		}

		return kr.reader.CommitMessages(ctx, msg)
	}

	return retry.ExecuteWithRetries(ctx, act)
}
