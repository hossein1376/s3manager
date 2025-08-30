package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/encrypt"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

type createObjectData struct {
	Bucket       string    `json:"bucket"`
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	Location     string    `json:"location"`
	LastModified time.Time `json:"last_modified"`
}

func (h *Handler) CreateObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := chi.URLParam(r, BucketName)
	if bucketName == "" {
		resp := serde.Response{Message: "bucket name is required"}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	err := r.ParseMultipartForm(32 << 20) // 32 Mb
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("parsing multipart form: %w", err),
		)
		return
	}
	path := r.FormValue("path")
	file, _, err := r.FormFile("file")
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("getting file from form: %w", err),
		)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.ErrorContext(ctx, "closing file", slog.Any("error", err))
		}
	}()

	opts := minio.PutObjectOptions{ContentType: "application/octet-stream"}
	if h.cfg.SSE.Type == "KMS" {
		opts.ServerSideEncryption, _ = encrypt.NewSSEKMS(h.cfg.SSE.Key, nil)
	}
	if h.cfg.SSE.Type == "SSE" {
		opts.ServerSideEncryption = encrypt.NewSSE()
	}
	if h.cfg.SSE.Type == "SSE-C" {
		opts.ServerSideEncryption, err = encrypt.NewSSEC([]byte(h.cfg.SSE.Key))
		if err != nil {
			serde.InternalErrWrite(
				ctx, w, fmt.Errorf("setting SSE-C key: %w", err),
			)
			return
		}
	}

	obj, err := h.s3.PutObject(ctx, bucketName, path, file, -1, opts)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("putting object: %w", err))
		return
	}

	resp := serde.Response{Data: createObjectData{
		Bucket:       obj.Bucket,
		Key:          obj.Key,
		Size:         obj.Size,
		Location:     obj.Location,
		LastModified: obj.LastModified,
	}}
	serde.WriteJson(ctx, w, http.StatusCreated, resp)
}
