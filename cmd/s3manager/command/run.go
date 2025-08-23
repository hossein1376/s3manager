package command

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/minio/minio-go/v7"

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

	s3, err := minio.New(cfg.S3.Endpoint, cfg.MinioOpts())
	if err != nil {
		return fmt.Errorf("creating s3 client: %w", err)
	}

	srv, err := handlers.NewServer(cfg, s3)
	if err != nil {
		return fmt.Errorf("new server: %w", err)
	}

	errCh := make(chan error)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	go func() {
		slog.InfoContext(ctx, "starting server", slog.String("address", srv.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-signalCh:
		return srv.Shutdown(ctx)
	}
}
