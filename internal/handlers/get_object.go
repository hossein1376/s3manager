package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
)

func (h *Handler) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	objectName := r.PathValue("object")
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
	v.Check(
		"key",
		validator.Case{
			Cond: validator.Empty(objectName), Msg: "object name is required",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	object, ct, err := h.service.GetObject(ctx, bucketName, objectName)
	if err != nil {
		grape.RespondFromErr(ctx, w, fmt.Errorf("getting object: %w", err))
		return
	}
	defer object.Close()

	contentType := "application/octet-stream"
	if ct != nil {
		contentType = *ct
	}
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", objectName),
	)
	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, object)
	if err != nil {
		grape.RespondFromErr(ctx, w, fmt.Errorf("copying object: %w", err))
		return
	}

	grape.WriteJson(ctx, w)
}
