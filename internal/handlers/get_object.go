package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	bucketName := strings.TrimSpace(vars["bucketName"])
	objectName := strings.TrimSpace(vars["objectName"])
	if bucketName == "" || objectName == "" {
		resp := serde.Response{
			Message: "bucket name and object name must be specified",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	object, err := h.s3.GetObject(
		ctx, bucketName, objectName, minio.GetObjectOptions{},
	)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("getting object: %w", err))
		return
	}

	if !h.cfg.S3.DisableForceDownload {
		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", objectName),
		)
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	_, err = io.Copy(w, object)
	if err != nil {
		serde.ExtractAndWrite(
			ctx, w, fmt.Errorf("copying object to response writer: %w", err),
		)
		return
	}
}
