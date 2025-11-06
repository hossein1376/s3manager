package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
	"github.com/hossein1376/s3manager/internal/model"
)

func (h *Handler) ListObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	v := validator.New()
	v.Check(
		"bucket",
		validator.Case{
			Cond: validator.Empty(bucketName), Msg: "bucket name is required",
		},
		validator.Case{
			Cond: validator.LengthMin(bucketName, 3),
			Msg:  "Bucket name cannot be shorter than 3 characters",
		},
		validator.Case{
			Cond: validator.Contains(bucketName, "/"),
			Msg:  "Bucket name cannot contain invalid characters",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	query := r.URL.Query()
	filter := query.Get("filter")
	token := query.Get("token")
	path := query.Get("path")

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
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	opts := model.ListObjectsOption{
		Path:   path,
		Filter: filter,
	}
	if token != "" {
		opts.ContinuationToken = &token
	}
	list, next, err := h.service.ListObjects(ctx, bucketName, int32(count), opts)
	if err != nil {
		grape.RespondFromErr(ctx, w, err)
		return
	}
	resp := listObjectsResponse{List: list, NextToken: next}
	grape.WriteJson(ctx, w, grape.WithData(resp))
}

type listObjectsResponse struct {
	List      []model.Object `json:"list"`
	NextToken *string        `json:"next_token,omitempty"`
}
