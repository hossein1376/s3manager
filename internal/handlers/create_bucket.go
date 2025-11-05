package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateBucketRequest
	err := serde.ReadJson(r, &req)
	if err != nil {
		resp := serde.Response{
			Message: fmt.Sprintf("decoding body JSON: %s", err),
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	if len(req.Name) < 3 {
		resp := serde.Response{
			Message: "Bucket name cannot be shorter than 3 characters",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	if strings.Contains(req.Name, "/") {
		resp := serde.Response{
			Message: "Bucket name cannot contain invalid characters",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	err = h.service.CreateBucket(ctx, req.Name)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, err)
		return
	}

	serde.WriteJson(ctx, w, http.StatusNoContent, nil)
}

type CreateBucketRequest struct {
	Name string `json:"name"`
}
