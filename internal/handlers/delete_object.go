package handlers

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) DeleteObjectHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucketName := r.PathValue("bucket")
	objectName := r.PathValue("object")
	if bucketName == "" || objectName == "" {
		resp := serde.Response{
			Message: "bucket name and object name must be specified",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	err := h.service.DeleteObject(ctx, bucketName, objectName)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("removing object: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}
