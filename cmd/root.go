package cmd

import (
	"context"
	"fmt"
	"os"
	sig "os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "saml-proxy",
		Short: "SAML Proxy",
		Long:  "SAML Proxy",
	}
)

func Execute() {
	rootCmd.Execute()
}

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
