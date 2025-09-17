package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

type createObjectData struct {
	Bucket       string    `json:"bucket"`
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	Location     string    `json:"location"`
	LastModified time.Time `json:"last_modified"`
}

func (h *Handler) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	objectKey := r.FormValue("key")
	if bucketName == "" || objectKey == "" {
		resp := serde.Response{Message: "bucket name and object key are required"}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	err := r.ParseMultipartForm(h.cfg.S3.MaxSizeBytes)
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("parsing multipart form: %w", err),
		)
		return
	}
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
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		serde.InternalErrWrite(ctx, w, fmt.Errorf("detecting mimetype: %w", err))
		return
	}

	obj, err := h.service.PutObject(
		ctx, bucketName, objectKey, mimeType.String(), file,
	)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("putting object: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusCreated, serde.Response{Data: obj})
}
