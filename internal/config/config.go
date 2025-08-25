package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

func (cfg *Config) MinioOpts() *minio.Options {
	opts := &minio.Options{
		Secure: !cfg.ConnectionOptions.SkipSSL,
	}
	if cfg.IAM.Enabled {
		opts.Creds = credentials.NewIAM(cfg.IAM.Endpoint)
	} else {
		opts.Creds = credentials.NewStatic(
			cfg.S3.AccessKeyID,
			cfg.S3.SecretAccessKey,
			"", // token
			credentials.SignatureType(cfg.SignatureType),
		)
	}

	if r := cfg.S3.Region; r != "" {
		opts.Region = r
	}
	if opts.Secure && cfg.ConnectionOptions.SkipSSLVerification {
		opts.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}

	return opts
}

type Config struct {
	S3                S3                `yaml:"s3"`
	IAM               IAM               `yaml:"iam"`
	SignatureType     SignatureType     `yaml:"signature_type"`
	ConnectionOptions ConnectionOptions `yaml:"connection_options"`
	Server            Server            `yaml:"server"`
	SSE               SSE               `yaml:"sse"`
}

type S3 struct {
	Endpoint             string `yaml:"endpoint"`
	AccessKeyID          string `yaml:"access_key"`
	SecretAccessKey      string `yaml:"secret_access_key"`
	Region               string `yaml:"region"`
	DisableListRecursive bool   `yaml:"disable_list_recursive"`
	DisableDelete        bool   `yaml:"disable_delete"`
	DisableForceDownload bool   `json:"disable_force_download"`
}

type IAM struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

type SSE struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key"`
	Type    string `yaml:"type"`
}

type ConnectionOptions struct {
	SkipSSL             bool `yaml:"skip_ssl"`
	SkipSSLVerification bool `yaml:"skip_ssl_verification"`
}

type Server struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type SignatureType credentials.SignatureType

func (s *SignatureType) UnmarshalYAML(bytes []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(bytes))) {
	case "v2":
		*s = SignatureType(credentials.SignatureV2)
	case "v4":
		*s = SignatureType(credentials.SignatureV4)
	case "v4streaming":
		*s = SignatureType(credentials.SignatureV4Streaming)
	case "anonymous":
		*s = SignatureType(credentials.SignatureAnonymous)
	default:
		return fmt.Errorf("unknown signature type: %s", bytes)
	}
	return nil
}
