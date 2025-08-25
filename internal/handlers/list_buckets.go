package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

type listBucketsData struct {
	Buckets     []minio.BucketInfo
	AllowDelete bool
}

func (h *Handler) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buckets, err := h.s3.ListBuckets(ctx)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, err)
		return
	}

	t, err := template.ParseFS(
		h.templates, "layout.html.tmpl", "buckets.html.tmpl",
	)
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("parsing template files: %w", err),
		)
		return
	}

	data := listBucketsData{
		Buckets:     buckets,
		AllowDelete: !h.cfg.S3.DisableDelete,
	}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("executing template: %w", err),
		)
		return
	}
}
