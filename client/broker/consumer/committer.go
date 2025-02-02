package consumer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Commit commits the processed messages. The method also removes old messages and offsets.
func (r KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {

	uncomMsg, ok := r.uncommittedMessageService.GetMsgFromUncommited(msgUuid)
	if !ok {
		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	lastOffsetWithTimeStamp, ok := r.offsetService.GetOffset(uncomMsg.Msg.Partition)
	if !ok || lastOffsetWithTimeStamp.Value < uncomMsg.Msg.Offset {
		err := r.reader.CommitMessages(ctx, uncomMsg.Msg)
		if err != nil {
			return fmt.Errorf("failed to commit a message in the broker: %w", err)

		}
	}

	r.offsetService.AddOrUpdateOffset(uncomMsg.Msg.Partition, uncomMsg.Msg.Offset)
	r.uncommittedMessageService.DelMsgFromUncommitted(msgUuid)

	return nil
}
