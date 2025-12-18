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

func (h *Handler) DeleteObjectHandle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucketName := r.PathValue("bucket")
	objectName := r.PathValue("object")
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
			Cond: validator.Not(validator.Empty(bucketName)),
			Msg:  "bucket name is required",
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
	v.Check(
		"key",
		validator.Case{
			Cond: !validator.Empty(objectName), Msg: "object name is required",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJSON(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	err = h.service.DeleteObject(ctx, bucketName, objectName, recursive)
	if err != nil {
		grape.ExtractFromErr(ctx, w, fmt.Errorf("removing object: %w", err))
		return
	}

	grape.WriteJSON(ctx, w, grape.WithStatus(http.StatusNoContent))
}
