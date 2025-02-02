package broker_data

import (
	"testing"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func TestRemoveOldMessages(t *testing.T) {
	msgService := NewUncommitedMessageService()
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
			testUuid := msgService.AddMsgToUncommitted(kafka.Message{})

			msgService.removeOldMessages(tt.input)
			_, ok := msgService.GetMsgFromUncommited(testUuid)
			if tt.want != ok {
				t.Errorf("have the message: %v, want: %v", ok, tt.want)
			}
		})
	}
}
