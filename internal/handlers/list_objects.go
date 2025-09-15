package handlers

import (
	"net/http"
	"strconv"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
	"github.com/hossein1376/s3manager/internal/model"
)

func (h *Handler) ListObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	if bucketName == "" {
		serde.WriteJson(
			ctx, w, http.StatusBadRequest, serde.Response{Message: "missing bucket name"},
		)
		return
	}
	query := r.URL.Query()
	filter := query.Get("filter")
	token := query.Get("token")
	countQuery := query.Get("count")

	var count int64 = 50
	if countQuery != "" {
		var err error
		count, err = strconv.ParseInt(countQuery, 10, 32)
		if err != nil {
			serde.WriteJson(
				ctx, w, http.StatusBadRequest, serde.Response{Message: "bad count"},
			)
			return
		}
	}

	opts := model.ListObjectsOption{}
	if token != "" {
		opts.ContinuationToken = &token
	}
	if filter != "" {
		opts.Filter = &filter
	}
	list, next, err := h.service.ListObjects(ctx, bucketName, int32(count), opts)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, err)
		return
	}
	resp := ListObjectsResponse{List: list, NextToken: next}
	serde.WriteJson(ctx, w, http.StatusOK, resp)
}

type ListObjectsResponse struct {
	List      []model.Object `json:"list"`
	NextToken *string        `json:"next_token,omitempty"`
}
