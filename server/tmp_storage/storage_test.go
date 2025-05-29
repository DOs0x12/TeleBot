package tmp_storage

import (
	"testing"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestRemoveOldObjs(t *testing.T) {
	msgService := NewTmpStorage()
	threshold := 1 * time.Hour
	tests := []struct {
		name  string
		input time.Duration
		want  bool
	}{
		{"Have a message", threshold, true},
		{"Have no message with the equal threshold", -threshold, false},
		{"Have no message with a greater threshold", -(threshold + 1*time.Millisecond), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testUuid := msgService.AddObj(kafka.Message{})

			msgService.delOldObjs(tt.input)
			_, ok := msgService.GetObj(testUuid)
			assert.Equal(t, tt.want, ok)
		})
	}
}
