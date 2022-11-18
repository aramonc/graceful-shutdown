package gracefulshutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

// CloserFunc is a function that accepts a context that can close
// a service or connection
type CloserFunc func(context.Context)

// Closer is an object that tracks shutdown functions and triggers
// them on a termination signal
type Closer struct {
	// Timeout defines how long the closer has to allow CloserFuncs to close the service
	Timeout time.Duration
	funcs   []CloserFunc
	ctx     context.Context
	stop    context.CancelFunc
}

func BuildCloser(pc context.Context, timeout time.Duration, signals ...os.Signal) (context.Context, *Closer) {
	ctx, stop := signal.NotifyContext(pc, signals...)

	return ctx, &Closer{
		Timeout: timeout,
		stop:    stop,
		ctx:     ctx,
	}
}

// Track adds
func (c *Closer) Track(f CloserFunc) {
	c.funcs = append(c.funcs, f)
}

func (c *Closer) Wait() error {
	<-c.ctx.Done()

	c.Close()

	return nil
}

func (c *Closer) Close() {
	c.stop()

	wg := &sync.WaitGroup{}
	wg.Add(len(c.funcs))

	for i, shutdown := range c.funcs {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), c.Timeout)

		go func(s CloserFunc, ctx context.Context, c context.CancelFunc, wg *sync.WaitGroup, i int) {
			s(ctx)
			c()
			wg.Done()
		}(shutdown, shutdownCtx, cancel, wg, i)
	}

	wg.Wait()
}
