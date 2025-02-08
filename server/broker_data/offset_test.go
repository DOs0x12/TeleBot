package broker_data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRemoveOldOffsets(t *testing.T) {
	offsetService := NewOffsetService()
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
			partVal := 1
			var offsetVal int64 = 1
			offsetService.AddOrUpdateOffset(partVal, offsetVal)

			offsetService.removeOldOffsets(tt.input)

			_, ok := offsetService.GetOffset(partVal)
			assert.Equal(t, tt.want, ok)
		})
	}
}
