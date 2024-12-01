package broker

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

func TestRemoveOldOffcets(t *testing.T) {
	threshold := 1 * time.Hour
	tests := []struct {
		name  string
		input time.Duration
		want  bool
	}{
		{"Have an offcet", threshold, true},
		{"Have no offcet with the equal threashold", -threshold, false},
		{"Have no offcet with a greater threashold", -(threshold + 1*time.Millisecond), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOffcets := make(map[int]offcetWithTimeStamp)
			offcetVal := 1
			testOffcets[offcetVal] = offcetWithTimeStamp{value: 1, timeStamp: time.Now()}

			removeOldOffcets(testOffcets, tt.input)
			_, ok := testOffcets[offcetVal]
			if tt.want != ok {
				t.Errorf("have the offcet: %v, want: %v", ok, tt.want)
			}
		})
	}
}
