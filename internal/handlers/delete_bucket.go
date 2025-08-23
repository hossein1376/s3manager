package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	if h.cfg.S3.DisableDelete {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	bucketName := mux.Vars(r)["bucketName"]
	err := h.s3.RemoveBucket(r.Context(), bucketName)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("removing bucket: %w", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
