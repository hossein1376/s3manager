package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
	"github.com/hossein1376/s3manager/pkg/icons"
)

var (
	bucketsRegEx = regexp.MustCompile(`/buckets/([^/]*)/?(.*)`)
)

type objectWithIcon struct {
	Icon         string
	DisplayName  string
	Key          string
	Size         int64
	Owner        string
	IsFolder     bool
	LastModified time.Time
}

type viewBucketData struct {
	BucketName  string
	CurrentPath string
	AllowDelete bool
	Paths       []string
	Objects     []objectWithIcon
}

func (h *Handler) ViewBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	matches := bucketsRegEx.FindStringSubmatch(r.RequestURI)
	if len(matches) != 3 {
		resp := serde.Response{
			Message: fmt.Sprintf("Invalid bucket URI: %s", r.RequestURI),
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}
	bucketName := matches[1]
	prefix := matches[2]

	doneCh := make(chan struct{})
	defer close(doneCh)

	var objs []objectWithIcon
	opts := minio.ListObjectsOptions{
		Recursive: !h.cfg.S3.DisableListRecursive,
		Prefix:    prefix,
	}

	objectCh := h.s3.ListObjects(ctx, bucketName, opts)
	for object := range objectCh {
		if object.Err != nil {
			serde.ExtractAndWrite(
				ctx, w, fmt.Errorf("listing objects: %w", object.Err),
			)
			return
		}

		obj := objectWithIcon{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			Owner:        object.Owner.DisplayName,
			Icon:         icons.Detect(object.Key).String(),
			IsFolder:     strings.HasSuffix(object.Key, "/"),
			DisplayName:  strings.TrimSuffix(strings.TrimPrefix(object.Key, prefix), "/"),
		}
		objs = append(objs, obj)
	}
	data := viewBucketData{
		BucketName:  bucketName,
		Objects:     objs,
		AllowDelete: !h.cfg.S3.DisableDelete,
		Paths: slices.DeleteFunc(
			strings.Split(prefix, "/"),
			func(s string) bool { return s == "" },
		),
		CurrentPath: prefix,
	}

	t, err := template.ParseFS(h.templates, "layout.html.tmpl", "bucket.html.tmpl")
	if err != nil {
		serde.InternalErrWrite(
			ctx, w, fmt.Errorf("parsing template files: %w", err),
		)
		return
	}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		serde.InternalErrWrite(ctx, w, fmt.Errorf("executing template: %w", err))
		return
	}
}
