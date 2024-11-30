package broker

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func TestRemoveOldMessages(t *testing.T) {
	testMessages := make(map[uuid.UUID]uncommittedMessage)
	testUuid := uuid.New()
	testMessages[testUuid] = uncommittedMessage{msg: kafka.Message{}, timeStamp: time.Now()}
	threshold := 1 * time.Hour
	removeOldMessages(testMessages, threshold)
	want := true
	_, ok := testMessages[testUuid]
	if want != ok {
		t.Error("No message in the map, want: the message is")
	}

	testOffcet := -threshold
	testMessages[testUuid] = uncommittedMessage{msg: kafka.Message{}, timeStamp: time.Now().Add(testOffcet)}
	removeOldMessages(testMessages, threshold)
	_, ok = testMessages[testUuid]
	if want == ok {
		t.Error("The message in the map, want: no message")
	}

	testOffcet = -(threshold + 1*time.Millisecond)
	testMessages[testUuid] = uncommittedMessage{msg: kafka.Message{}, timeStamp: time.Now().Add(testOffcet)}
	removeOldMessages(testMessages, threshold)
	_, ok = testMessages[testUuid]
	if want == ok {
		t.Error("The message in the map, want: no message")
	}
}
