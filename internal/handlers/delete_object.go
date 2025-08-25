package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) DeleteObjectHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.cfg.S3.DisableDelete {
		serde.WriteJson(ctx, w, http.StatusForbidden, nil)
		return
	}

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

	err := h.s3.RemoveObject(
		ctx, bucketName, objectName, minio.RemoveObjectOptions{},
	)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("removing object: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}
