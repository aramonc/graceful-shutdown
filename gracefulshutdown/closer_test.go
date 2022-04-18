package gracefulshutdown_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aramonc/graceful-shutdown/gracefulshutdown"
)

func TestContextHasToBeTerminable(t *testing.T) {
	closer := &gracefulshutdown.Closer{Timeout: 200 * time.Millisecond}

	err := closer.Wait(context.Background())

	assert.Error(t, err)
}

func TestThatClosesMultipleConnections(t *testing.T) {
	expectedClosed := 3
	actualClosed := 0
	didClose := make(chan int, 3)
	closer := &gracefulshutdown.Closer{Timeout: 200 * time.Millisecond}

	closer.Track(func(ctx context.Context) {
		// can ignore context if it is not needed
		didClose <- 1
	})

	closer.Track(func(ctx context.Context) {
		// can wait for the timeout
		<-ctx.Done()
		didClose <- 1
	})

	closer.Track(func(ctx context.Context) {
		// can wait the context to be cancelled
		go func(ctx context.Context, closed chan int) {
			<-ctx.Done()
			closed <- 1
		}(ctx, didClose)
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := closer.Wait(ctx)

	require.NoError(t, err)

	timedOut := false
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	for !timedOut && actualClosed != expectedClosed {
		select {
		case closed := <-didClose:
			actualClosed += closed
		case <-timer.C:
			timedOut = true
		}
	}

	assert.Equal(t, expectedClosed, actualClosed, "did not closed all services")
	assert.False(t, timedOut, "timed out")
}
