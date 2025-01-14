package consumer

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func TestRemoveOldMessages(t *testing.T) {
	threshold := 1 * time.Hour
	tests := []struct {
		name  string
		input time.Duration
		want  bool
	}{
		{"Have a message", threshold, true},
		{"Have no message with the equal threashold", -threshold, false},
		{"Have no message with a greater threashold", -(threshold + 1*time.Millisecond), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMessages := make(map[uuid.UUID]uncommittedMessage)
			testUuid := uuid.New()
			testMessages[testUuid] = uncommittedMessage{msg: kafka.Message{}, timeStamp: time.Now()}

			removeOldMessages(testMessages, tt.input)
			_, ok := testMessages[testUuid]
			if tt.want != ok {
				t.Errorf("have the message: %v, want: %v", ok, tt.want)
			}
		})
	}
}

func TestRemoveOldOffsets(t *testing.T) {
	threshold := 1 * time.Hour
	tests := []struct {
		name  string
		input time.Duration
		want  bool
	}{
		{"Have an offset", threshold, true},
		{"Have no offset with the equal threashold", -threshold, false},
		{"Have no offset with a greater threashold", -(threshold + 1*time.Millisecond), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOffsets := make(map[int]offsetWithTimeStamp)
			offsetVal := 1
			testOffsets[offsetVal] = offsetWithTimeStamp{value: 1, timeStamp: time.Now()}

			removeOldOffsets(testOffsets, tt.input)
			_, ok := testOffsets[offsetVal]
			if tt.want != ok {
				t.Errorf("have the offset: %v, want: %v", ok, tt.want)
			}
		})
	}
}
