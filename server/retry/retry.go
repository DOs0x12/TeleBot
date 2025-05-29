package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type action func(ctx context.Context) error

// ExecuteWithRetries executes a function with retries using a defined retry count and delay between executes.
func ExecuteWithRetries(ctx context.Context, act action, retryCnt int, delay time.Duration) error {
	var err error

	for i := range retryCnt {
		if ctx.Err() != nil {
			return nil
		}

		err = act(ctx)
		if err == nil {
			return nil
		}

		logrus.Errorf("an error occurs: %v\nRetry %v after %v", err, i, delay)
		waitWithContext(ctx, delay)
	}

	return fmt.Errorf("retries are exceeded: %w", err)
}

func waitWithContext(ctx context.Context, waitTime time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(waitTime):
	}
}
