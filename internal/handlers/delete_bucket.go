package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.cfg.S3.DisableDelete {
		serde.WriteJson(ctx, w, http.StatusForbidden, nil)
		return
	}

	bucketName := strings.TrimSpace(chi.URLParam(r, BucketName))
	if bucketName == "" {
		resp := serde.Response{Message: "bucket name is required"}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	err := h.s3.RemoveBucket(ctx, bucketName)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("removing bucket: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}
