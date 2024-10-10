package interruption

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WatchForInterruption(cancels ...context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		for _, cancel := range cancels {
			cancel()
		}
	}()
}
