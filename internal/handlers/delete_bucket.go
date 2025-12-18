package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/validator"
)

func (h *Handler) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	recursive, err := grape.Query(r.URL.Query(), "recursive", strconv.ParseBool)
	if err != nil {
		if !errors.Is(err, grape.ErrMissingQuery) {
			grape.ExtractFromErr(
				ctx, w, errs.BadRequest(errs.WithMsg("Invalid recursive param")),
			)
			return
		}
		recursive = false
	}
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
			Cond: validator.Not(validator.Contains(bucketName, "/")),
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

	err = h.service.DeleteBucket(ctx, bucketName, recursive)
	if err != nil {
		grape.ExtractFromErr(ctx, w, fmt.Errorf("removing bucket: %w", err))
		return
	}

	grape.WriteJSON(ctx, w, grape.WithStatus(http.StatusNoContent))
}
