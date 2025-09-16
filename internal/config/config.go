package config

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

func New(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	cfg := &Config{}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// validate

	return cfg, nil
}

type Config struct {
	S3     S3     `yaml:"s3"`
	Server Server `yaml:"server"`
}

type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
}

type Server struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}
