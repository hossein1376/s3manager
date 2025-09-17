package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

func New(path string) (Config, error) {
	cfg := Config{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return defaults(), nil
	default:
		return cfg, fmt.Errorf("reading file: %w", err)
	}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal: %w", err)
	}

	return cfg, nil
}

func defaults() Config {
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	if accessKey == "" {
		accessKey = os.Getenv("MINIO_ACCESS_KEY")
	}
	secretKey := os.Getenv("AWS_SECRET_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("MINIO_SECRET_KEY")
	}
	return Config{
		S3: S3{
			Endpoint:        "http://127.0.0.1:9000",
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			Region:          "auto",
		},
		Server: Server{
			Address:      "0.0.0.0:8080",
			ReadTimeout:  2 * time.Minute,
			WriteTimeout: 1 * time.Minute,
			DisableUI:    false,
		},
		Logger: Logger{
			Level: slog.LevelInfo,
		},
		IsDefault: true,
	}
}

type Config struct {
	S3        S3     `yaml:"s3"`
	Server    Server `yaml:"server"`
	Logger    Logger `yaml:"logger"`
	IsDefault bool   `yaml:"-"`
}

type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access-key"`
	SecretAccessKey string `yaml:"secret-access-key"`
	Region          string `yaml:"region"`
}

type Server struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout"`
	DisableUI    bool          `yaml:"disable-ui"`
}

type Logger struct {
	Level slog.Level `yaml:"level"`
}
