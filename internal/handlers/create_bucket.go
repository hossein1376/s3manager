package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var bucket minio.BucketInfo
	err := json.NewDecoder(r.Body).Decode(&bucket)
	if err != nil {
		resp := serde.Response{
			Message: fmt.Sprintf("decoding body JSON: %s", err),
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	name := bucket.Name
	if len(name) < 3 {
		resp := serde.Response{
			Message: "Bucket name cannot be shorter than 3 characters",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	err = h.s3.MakeBucket(ctx, bucket.Name, minio.MakeBucketOptions{})
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("making bucket: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}
