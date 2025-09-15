package handlers

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucketName := r.PathValue("bucket")
	if bucketName == "" {
		resp := serde.Response{Message: "bucket name is required"}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	err := h.service.DeleteBucket(ctx, bucketName)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("removing bucket: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}
