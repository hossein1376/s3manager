package handlers

import (
	"fmt"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
	"github.com/hossein1376/s3manager/pkg/slogger"
)

func (h *Handler) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	objectKey := r.FormValue("key")
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
			Cond: validator.Empty(objectKey), Msg: "object name is required",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	err := r.ParseMultipartForm(h.cfg.S3.MaxSizeBytes)
	if err != nil {
		grape.RespondFromErr(
			ctx, w, fmt.Errorf("parsing multipart form: %w", err),
		)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		grape.RespondFromErr(
			ctx, w, fmt.Errorf("getting file from form: %w", err),
		)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			slogger.Error(ctx, "closing file", slogger.Err("error", err))
		}
	}()
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		grape.RespondFromErr(ctx, w, fmt.Errorf("detecting mimetype: %w", err))
		return
	}

	obj, err := h.service.PutObject(
		ctx, bucketName, objectKey, mimeType.String(), file,
	)
	if err != nil {
		grape.RespondFromErr(ctx, w, fmt.Errorf("putting object: %w", err))
		return
	}

	grape.WriteJson(
		ctx,
		w,
		grape.WithStatus(http.StatusCreated),
		grape.WithData(grape.Response{Data: obj}),
	)
}
