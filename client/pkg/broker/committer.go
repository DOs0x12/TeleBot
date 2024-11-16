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
	uncomMsg, ok := r.uncommittedMessages[msgUuid]
	if !ok {
		r.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	r.mu.Unlock()

	err := r.reader.CommitMessages(ctx, uncomMsg.msg)
	if err != nil {
		return fmt.Errorf("can not commit a message in the broker: %w", err)

	}

	r.mu.Lock()
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
