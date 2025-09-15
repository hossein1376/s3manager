package handlers

import (
	"net/http"
	"strconv"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
	"github.com/hossein1376/s3manager/internal/model"
)

func (h *Handler) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	filter := query.Get("filter")
	token := query.Get("token")
	countStr := query.Get("count")

	var count int64 = 50
	if countStr != "" {
		var err error
		count, err = strconv.ParseInt(countStr, 10, 32)
		if err != nil {
			serde.ExtractAndWrite(ctx, w, err) //TODO
			return
		}
	}
	opts := model.ListBucketsOptions{}
	if filter != "" {
		opts.Filter = &filter
	}
	if token != "" {
		opts.ContinuationToken = &token
	}

	buckets, next, err := h.service.ListBuckets(ctx, int32(count), opts)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, err)
		return
	}
	resp := ListBucketsResponse{Buckets: buckets, NextToken: next}
	serde.WriteJson(ctx, w, http.StatusOK, resp)
}

type ListBucketsResponse struct {
	Buckets   []model.Bucket `json:"buckets"`
	NextToken *string        `json:"next_token,omitempty"`
}
