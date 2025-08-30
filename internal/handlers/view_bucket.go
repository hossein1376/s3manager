package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
	"github.com/hossein1376/s3manager/pkg/icons"
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
	Filter      string
	Count       string
	Recursive   bool
	BucketName  string
	CurrentPath string
	AllowDelete bool
	Paths       []string
	Objects     []objectWithIcon
}

func (h *Handler) ViewBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	filter := query.Get("filter")
	recursive := query.Get("recursive") == "true"
	countQuery := query.Get("count")

	count := 50
	if countQuery != "" {
		var err error
		count, err = strconv.Atoi(countQuery)
		if err != nil {
			serde.WriteJson(
				ctx, w, http.StatusBadRequest, serde.Response{Message: "bad count"},
			)
			return
		}
	}

	var bucketName, path string
	switch args := strings.SplitN(chi.URLParam(r, "*"), "/", 2); len(args) {
	case 1:
		bucketName = args[0]
	case 2:
		bucketName = args[0]
		path = strings.TrimPrefix(args[1], "/")
	default:
		serde.WriteJson(
			ctx,
			w,
			http.StatusBadRequest,
			serde.Response{Message: "bad path parameter"},
		)
		return
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	objs := make([]objectWithIcon, 0, count)
	opts := minio.ListObjectsOptions{
		Recursive: recursive,
		Prefix:    path + filter,
	}

	objectCh := h.s3.ListObjects(ctx, bucketName, opts)
	for object := range objectCh {
		if object.Err != nil {
			serde.ExtractAndWrite(
				ctx, w, fmt.Errorf("listing objects: %w", object.Err),
			)
			return
		}

		if count != 0 && len(objs) > count {
			break
		}

		obj := objectWithIcon{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			Owner:        object.Owner.DisplayName,
			Icon:         icons.Detect(object.Key).String(),
			IsFolder:     strings.HasSuffix(object.Key, "/"),
			DisplayName:  strings.TrimPrefix(object.Key, path),
		}
		objs = append(objs, obj)
	}
	paths := slices.DeleteFunc(
		strings.Split(path, "/"),
		func(s string) bool { return s == "" },
	)
	current := "/"
	if len(paths) > 1 {
		current = paths[len(paths)-1]
	}

	data := viewBucketData{
		Filter:      filter,
		Recursive:   recursive,
		BucketName:  bucketName,
		Objects:     objs,
		AllowDelete: !h.cfg.S3.DisableDelete,
		Paths:       paths,
		CurrentPath: current,
		Count:       countQuery,
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
