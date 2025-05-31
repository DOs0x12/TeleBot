package consumer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Commit commits the processed messages. The method also removes old messages and offsets.
func (r KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {

	stObj, ok := r.uncomMsgStorage.GetObj(msgUuid)
	if !ok {
		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	uncomMsg := stObj.Obj.(kafka.Message)
	lastOffsetWithTimeStamp, ok := r.offsetService.GetOffset(uncomMsg.Partition)
	if !ok || lastOffsetWithTimeStamp.Value < uncomMsg.Offset {
		err := r.reader.CommitMessages(ctx, uncomMsg)
		if err != nil {
			return fmt.Errorf("failed to commit a message in the broker: %w", err)

		}
	}

	r.offsetService.AddOrUpdateOffset(uncomMsg.Partition, uncomMsg.Offset)
	r.uncomMsgStorage.DelObj(msgUuid)

	return nil
}
