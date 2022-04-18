package gracefulshutdown

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CloserFunc is a function that accepts a context that can close
// a service or connection
type CloserFunc func(context.Context)

// Closer is an object that tracks shutdown functions and triggers
// them on ctx.Close()
type Closer struct {
	// Timeout defines how long the closer has to allow CloserFuncs to close the service
	Timeout time.Duration
	funcs   []CloserFunc
}

// Track adds
func (c *Closer) Track(f CloserFunc) {
	c.funcs = append(c.funcs, f)
}

func (c *Closer) Wait(ctx context.Context) error {
	doneBus := ctx.Done()
	if doneBus == nil {
		return errors.New("context is not terminable, cannot wait")
	}

	<-doneBus

	wg := &sync.WaitGroup{}
	wg.Add(len(c.funcs))

	for i, shutdown := range c.funcs {
		shutdownCtx, cancel := context.WithTimeout(ctx, c.Timeout)

		go func(s CloserFunc, ctx context.Context, c context.CancelFunc, wg *sync.WaitGroup, i int) {
			s(ctx)
			c()
			wg.Done()
		}(shutdown, shutdownCtx, cancel, wg, i)
	}

	wg.Wait()

	return nil
}
