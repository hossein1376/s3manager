package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucketName := r.PathValue("bucket")
	objectName := r.PathValue("object")
	if bucketName == "" || objectName == "" {
		resp := serde.Response{
			Message: "bucket name and object name must be specified",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	object, ct, err := h.service.GetObject(ctx, bucketName, objectName)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("getting object: %w", err))
		return
	}
	defer object.Close()
	var contentType string
	if ct != nil {
		contentType = *ct
	} else {
		contentType = "application/octet-stream"
	}

	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", objectName),
	)
	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, object)
	if err != nil {
		serde.InternalErrWrite(ctx, w, fmt.Errorf("copying object: %w", err))
		return
	}

	serde.WriteJson(ctx, w, http.StatusOK, nil)
}
