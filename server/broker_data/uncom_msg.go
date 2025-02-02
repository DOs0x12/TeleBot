package broker_data

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

// UncommitedMessageService stores uncommited messages of a broker.
type UncommitedMessageService struct {
	mu                  *sync.Mutex
	uncommittedMessages map[uuid.UUID]uncommittedMessage
}

type uncommittedMessage struct {
	Msg       kafka.Message
	timeStamp time.Time
}

// NewUncommitedMessageService creates a new NewUncommitedMessageService instance.
func NewUncommitedMessageService() UncommitedMessageService {
	return UncommitedMessageService{mu: &sync.Mutex{}, uncommittedMessages: make(map[uuid.UUID]uncommittedMessage)}
}

// StartCleanupUncommittedMessages starts a worker for cleaning up the old messages.
func (s UncommitedMessageService) StartCleanupUncommittedMessages(ctx context.Context) {
	go s.delOldUncommittedMessagesWithPeriod(ctx)
}

func (s UncommitedMessageService) delOldUncommittedMessagesWithPeriod(ctx context.Context) {
	t := time.NewTicker(1 * time.Hour)
	threshold := 48 * time.Hour
	for {
		select {
		case <-ctx.Done():
			t.Stop()

			return
		case <-t.C:
			s.removeOldMessages(threshold)
		}
	}
}

// AddMsgToUncommitted adds or updates a message to others.
func (s UncommitedMessageService) AddMsgToUncommitted(msg kafka.Message) uuid.UUID {
	msgUuid := uuid.New()
	s.mu.Lock()
	s.uncommittedMessages[msgUuid] = uncommittedMessage{Msg: msg, timeStamp: time.Now()}
	s.mu.Unlock()

	return msgUuid
}

// DelMsgFromUncommitted deletes a single message from the set by UUID.
func (s UncommitedMessageService) DelMsgFromUncommitted(msgUuid uuid.UUID) {
	s.mu.Lock()
	delete(s.uncommittedMessages, msgUuid)
	s.mu.Unlock()
}

// GetMsgFromUncommited gets a message by UUID.
func (s UncommitedMessageService) GetMsgFromUncommited(msgUuid uuid.UUID) (uncommittedMessage, bool) {
	s.mu.Lock()
	msg, ok := s.uncommittedMessages[msgUuid]
	s.mu.Unlock()
	return msg, ok
}

func (s UncommitedMessageService) removeOldMessages(threshold time.Duration) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for msgUuid, procMsg := range s.uncommittedMessages {
		if procMsg.timeStamp.Add(threshold).Before(now) {
			delete(s.uncommittedMessages, msgUuid)
		}
	}
}
