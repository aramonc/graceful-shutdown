package gracefulshutdown_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aramonc/graceful-shutdown/gracefulshutdown"
)

func TestItWaitsForSignal(t *testing.T) {
	ctx := gracefulshutdown.Listen(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	timedOut := false
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	for !timedOut {
		select {
		case <-ctx.Done():
			assert.False(t, timedOut, "handled signal")
			return
		case <-timer.C:
			timedOut = true
		}
	}

	assert.True(t, timedOut, "did not handle signal")
}
