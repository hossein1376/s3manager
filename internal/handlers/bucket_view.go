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

	"github.com/hossein1376/s3manager/pkg/icons"
)

var (
	bucketsRegEx = regexp.MustCompile(`/buckets/([^/]*)/?(.*)`)
)

type objectWithIcon struct {
	Key          string
	Size         int64
	LastModified time.Time
	Owner        string
	Icon         string
	IsFolder     bool
	DisplayName  string
}

type pageData struct {
	BucketName  string
	Objects     []objectWithIcon
	AllowDelete bool
	Paths       []string
	CurrentPath string
}

func (h *Handler) BucketViewHandler(w http.ResponseWriter, r *http.Request) {
	matches := bucketsRegEx.FindStringSubmatch(r.RequestURI)
	if len(matches) != 3 {
		// TODO ?
	}
	bucketName := matches[1]
	prefix := matches[2]

	var objs []objectWithIcon
	doneCh := make(chan struct{})
	defer close(doneCh)
	opts := minio.ListObjectsOptions{
		Recursive: !h.cfg.S3.DisableListRecursive,
		Prefix:    prefix,
	}
	objectCh := h.s3.ListObjects(r.Context(), bucketName, opts)
	for object := range objectCh {
		if object.Err != nil {
			handleHTTPError(w, fmt.Errorf("listing objects: %w", object.Err))
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
	data := pageData{
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
		handleHTTPError(w, fmt.Errorf("parsing template files: %w", err))
		return
	}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("executing template: %w", err))
		return
	}
}
