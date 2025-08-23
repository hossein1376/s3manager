package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/minio/minio-go/v7"
)

type listPageData struct {
	Buckets     []minio.BucketInfo
	AllowDelete bool
}

func (h *Handler) BucketsViewHandler(w http.ResponseWriter, r *http.Request) {

	buckets, err := h.s3.ListBuckets(r.Context())
	if err != nil {
		handleHTTPError(w, fmt.Errorf("listing buckets: %w", err))
		return
	}

	data := listPageData{
		Buckets:     buckets,
		AllowDelete: !h.cfg.S3.DisableDelete,
	}

	t, err := template.ParseFS(h.templates, "layout.html.tmpl", "buckets.html.tmpl")
	if err != nil {
		handleHTTPError(w, fmt.Errorf("parsing template files: %w", err))
		return
	}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("executing template: %w", err))
		return

	}
}
