package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

func (h *Handler) DeleteObjectHandle(w http.ResponseWriter, r *http.Request) {
	bucketName := mux.Vars(r)["bucketName"]
	objectName := mux.Vars(r)["objectName"]

	err := h.s3.RemoveObject(r.Context(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		handleHTTPError(w, fmt.Errorf("error removing object: %w", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
