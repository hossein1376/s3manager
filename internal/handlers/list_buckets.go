package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

type listBucketsData struct {
	Buckets     []minio.BucketInfo
	Filter      string
	Fold        bool
	AllowDelete bool
}

func (h *Handler) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	filter := query.Get("filter")

	buckets, err := h.s3.ListBuckets(ctx)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, err)
		return
	}
	if filter != "" {
		var filtered []minio.BucketInfo
		for _, bucket := range buckets {
			if strings.HasPrefix(bucket.Name, filter) {
				filtered = append(filtered, bucket)
			}
		}
		buckets = filtered
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
		Filter:      filter,
	}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("executing template: %w", err),
		)
		return
	}
}
