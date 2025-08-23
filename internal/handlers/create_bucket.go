package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
)

func (h *Handler) CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	var bucket minio.BucketInfo
	err := json.NewDecoder(r.Body).Decode(&bucket)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("decoding body JSON: %w", err))
		return
	}

	err = h.s3.MakeBucket(r.Context(), bucket.Name, minio.MakeBucketOptions{})
	if err != nil {
		handleHTTPError(w, fmt.Errorf("making bucket: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(bucket)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("encoding JSON: %w", err))
		return
	}
}
