package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type action func(ctx context.Context) error

func ExecuteWithRetries(ctx context.Context, act action, retryCnt int, waitTime time.Duration) error {
	var err error

	for i := 0; i < retryCnt; i++ {
		if ctx.Err() != nil {
			return nil
		}

		err = act(ctx)
		if err == nil {
			return nil
		}

		logrus.Errorf("an error occurs: %v\nRetry %v after %v", err, i, waitTime)
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
