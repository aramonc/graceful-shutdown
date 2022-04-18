package gracefulshutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var signals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
}

// Listen waits for a system signal and closes the
// context
func Listen(ctx context.Context) context.Context {
	shutdownCtx, _ := signal.NotifyContext(ctx, signals...)
	return shutdownCtx
}
