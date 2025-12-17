package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/s3manager/internal/model"
)

func (h *Handler) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	filter := query.Get("filter")
	token := query.Get("token")
	count, err := grape.Query(query, "count", grape.ParseInt[int32]())
	switch {
	case err == nil:
		// continue
	case errors.Is(err, grape.ErrMissingQuery):
		count = 50 // default value
	default:
		resp := grape.Response{
			Message: "Bad input", Data: fmt.Sprintf("parse count: %s", err),
		}
		grape.WriteJSON(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	opts := model.ListBucketsOptions{}
	if filter != "" {
		opts.Filter = &filter
	}
	if token != "" {
		opts.ContinuationToken = &token
	}

	buckets, next, err := h.service.ListBuckets(ctx, count, opts)
	if err != nil {
		grape.ExtractFromErr(ctx, w, err)
		return
	}
	resp := listBucketsResponse{Buckets: buckets, NextToken: next}
	grape.WriteJSON(ctx, w, grape.WithData(resp))
}

type listBucketsResponse struct {
	Buckets   []model.Bucket `json:"buckets"`
	NextToken *string        `json:"next_token,omitempty"`
}
