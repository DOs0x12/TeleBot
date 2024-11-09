package common

import (
	"context"
	"fmt"
	"time"
)

type action func(ctx context.Context) error

func ExecuteWithRetries(ctx context.Context, act action) error {
	retryNum := 10
	waitTime := 5 * time.Second
	var err error

	for i := 0; i < retryNum; i++ {
		if ctx.Err() != nil {
			return nil
		}

		err = act(ctx)

		waitWithContext(ctx, waitTime)
	}

	return fmt.Errorf("retries are exceeded: %w", err)
}

func waitWithContext(ctx context.Context, waitTime time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(waitTime):
	}
}
