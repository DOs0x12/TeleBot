package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r Receiver) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	r.mu.Lock()

	r.removeOldMessages()
	r.removeOldOffcets()
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

func (r Receiver) removeOldMessages() {
	now := time.Now()

	for msgUuid, procMsg := range r.uncommittedMessages {
		if procMsg.timeStamp.Add(48 * time.Hour).Before(now) {
			delete(r.uncommittedMessages, msgUuid)
		}
	}
}

func (r Receiver) removeOldOffcets() {
	now := time.Now()

	for part, offcet := range r.offcets {
		if offcet.timeStamp.Add(48 * time.Hour).Before(now) {
			delete(r.offcets, part)
		}
	}
}
