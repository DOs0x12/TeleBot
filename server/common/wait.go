package common

import (
	"context"
	"time"
)

func WaitWithContext(ctx context.Context, waitTime time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(waitTime):
	}
}
