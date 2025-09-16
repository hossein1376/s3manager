package command

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hossein1376/s3manager/internal/services"
	"github.com/hossein1376/s3manager/pkg/slogger"

	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/internal/handlers"
)

func Run() error {
	ctx := context.Background()

	var cfgPath string
	flag.StringVar(&cfgPath, "c", "assets/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.New(cfgPath)
	if err != nil {
		return fmt.Errorf("new config: %w", err)
	}

	logger := slogger.NewJSONLogger(slog.LevelDebug, os.Stdout)
	slog.SetDefault(logger)

	endpointURL, err := url.Parse(cfg.S3.Endpoint)
	if err != nil {
		return fmt.Errorf("parse s3 endpoint: %w", err)
	}
	addr, err := net.LookupIP(endpointURL.Hostname())
	switch {
	case err != nil:
		return fmt.Errorf("lookup ip: %w", err)
	case len(addr) == 0:
		return fmt.Errorf("hostname didn't resolve to an IP address")
	}
	endpointURL.Host = fmt.Sprintf("%s:%s", addr[0], endpointURL.Port())

	s3Client := s3.NewFromConfig(aws.Config{
		BaseEndpoint: aws.String(endpointURL.String()),
		Region:       cfg.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, "",
		),
		HTTPClient: nil,
	})

	srvc := services.New(s3Client)
	server, err := handlers.NewServer(cfg, srvc)
	if err != nil {
		return fmt.Errorf("new server: %w", err)
	}

	errCh := make(chan error)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	go func() {
		slog.InfoContext(ctx, "starting server", slog.String("address", server.Addr))
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-signalCh:
		return server.Shutdown(ctx)
	}
}
