package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/n-creativesystem/saml-proxy/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type webServer struct {
	port int
	cert string
	key  string
}

var (
	srvConf     webServer
	debug       bool
	samlConfig  string
	redis       string
	metricsName string
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			ctx := cmd.Context()
			if err := run(ctx); err != nil {
				logrus.Fatalln(err)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	runFlags := runCmd.PersistentFlags()
	runFlags.BoolVar(&debug, "debug", true, "debug mode")
	runFlags.StringVar(&samlConfig, "samlConfig", "saml", "saml configuration file name")
	runFlags.IntVar(&srvConf.port, "httpPort", 8080, "http port")
	runFlags.StringVar(&redis, "redis", "", "redis config")
	runFlags.StringVar(&srvConf.cert, "cert", "", "ssl certification file name")
	runFlags.StringVar(&srvConf.key, "key", "", "ssl key file name")
	runFlags.StringVar(&metricsName, "metrics-name", "saml_proxy", "prometheus metrics prefix name")
	_ = viper.BindPFlag("saml-config", runFlags.Lookup("samlConfig"))
	_ = viper.BindPFlag("metrics-name", runFlags.Lookup("metrics-name"))
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
		logrus.Infof("REST Server: %s", httpAddr)
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
	opts := []server.Option{
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
