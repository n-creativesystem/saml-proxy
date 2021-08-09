package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/n-creativesystem/saml-proxy/logger"
	"github.com/n-creativesystem/saml-proxy/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	Version  = "0.0.1"
	Revision = "rev"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: logger.TimestampFormat,
	})
	setEnvFlag()
	flag.Parse()
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if err := run(context.Background()); err != nil {
		logrus.Fatalln(err)
	}
}

func run(ctx context.Context) error {
	var (
		eg         *errgroup.Group
		httpLister net.Listener
	)
	defer func() {
		if httpLister != nil {
			httpLister.Close()
		}
	}()

	eg, ctx = errgroup.WithContext(ctx)

	var err error
	httpAddr := fmt.Sprintf(":%d", srvConf.port)
	httpLister, err = net.Listen("tcp", httpAddr)
	if err != nil {
		logrus.Fatalln(err)
	}
	eg.Go(func() error {
		logrus.Printf("REST Server: %s", httpAddr)
		return runRest(ctx, httpLister)
	})
	eg.Go(func() error {
		return signal(ctx)
	})
	eg.Go(func() error {
		<-ctx.Done()
		return ctx.Err()
	})
	return eg.Wait()
}

func runRest(ctx context.Context, li net.Listener) error {
	opts := []server.Options{
		server.WithSAMLConfig(samlConfig),
	}
	if debug {
		opts = append(opts, server.WithDebug)
	}
	if redis != "" {
		opts = append(opts, server.WithRedis(redis))
	}
	restSrv := server.New(opts...)
	httpServer := &http.Server{
		Handler:      restSrv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	errCh := make(chan error)
	go func() {
		if srvConf.cert != "" && srvConf.key != "" {
			if err := httpServer.ServeTLS(li, srvConf.cert, srvConf.key); err != http.ErrServerClosed {
				errCh <- err
			}
		} else {
			if err := httpServer.Serve(li); err != http.ErrServerClosed {
				errCh <- err
			}
		}
	}()
	select {
	case <-ctx.Done():
		cancelCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		return httpServer.Shutdown(cancelCtx)
	case err := <-errCh:
		return err
	}
}
