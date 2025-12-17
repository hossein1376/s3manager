package handlers

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
)

func (h *Handler) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	v := validator.New()
	v.Check(
		"bucket",
		validator.Case{
			Cond: !validator.Empty(bucketName), Msg: "bucket name is required",
		},
		validator.Case{
			Cond: validator.LengthMin(bucketName, 3),
			Msg:  "Bucket name cannot be shorter than 3 characters",
		},
		validator.Case{
			Cond: !validator.Contains(bucketName, "/"),
			Msg:  "Bucket name cannot contain invalid characters",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJSON(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	err := h.service.DeleteBucket(ctx, bucketName)
	if err != nil {
		grape.ExtractFromErr(ctx, w, fmt.Errorf("removing bucket: %w", err))
		return
	}

	grape.WriteJSON(ctx, w, grape.WithStatus(http.StatusNoContent))
}
