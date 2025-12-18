package command

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hossein1376/grape/slogger"
	"github.com/hossein1376/s3manager/internal/services"

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

	slogger.NewDefault(slogger.WithLevel(cfg.Logger.Level))
	if cfg.IsDefault {
		slog.Warn("using default configs, use -c flag to specify configuration file")
	}

	s3Client := s3.NewFromConfig(aws.Config{
		BaseEndpoint: aws.String(cfg.S3.Endpoint),
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
