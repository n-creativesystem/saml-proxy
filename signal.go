// +build !windows

package main

import (
	"context"
	"fmt"
	"os"
	sig "os/signal"
	"syscall"
)

func signal(ctx context.Context) error {
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGSTOP,
	}
	osNotify := make(chan os.Signal, 1)
	sig.Notify(osNotify, signals...)
	select {
	case <-ctx.Done():
		sig.Reset()
		return nil
	case s := <-osNotify:
		return fmt.Errorf("signal received: %v", s)
	}
}
