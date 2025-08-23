package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

func (h *Handler) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := mux.Vars(r)["bucketName"]
	objectName := mux.Vars(r)["objectName"]

	object, err := h.s3.GetObject(r.Context(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		handleHTTPError(w, fmt.Errorf("error getting object: %w", err))
		return
	}

	if !h.cfg.S3.DisableForceDownload {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	_, err = io.Copy(w, object)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("error copying object to response writer: %w", err))
		return
	}
}
