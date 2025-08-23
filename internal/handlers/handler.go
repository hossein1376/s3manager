package handlers

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/web"
)

type Handler struct {
	cfg       *config.Config
	s3        *minio.Client
	templates fs.FS
	statics   fs.FS
}

func NewServer(
	cfg *config.Config, s3 *minio.Client,
) (*http.Server, error) {
	templates, statics, err := web.Load()
	if err != nil {
		return nil, fmt.Errorf("loading web: %w", err)
	}
	h := &Handler{
		cfg:       cfg,
		s3:        s3,
		templates: templates,
		statics:   statics,
	}

	return &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      newRouter(h),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}, nil
}

func newRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	r.Handle("/", http.RedirectHandler(
		"/buckets", http.StatusPermanentRedirect),
	).Methods(http.MethodGet)

	r.PathPrefix("/static/").
		Handler(
			http.StripPrefix("/static/", http.FileServer(http.FS(h.statics))),
		).
		Methods(http.MethodGet)

	r.HandleFunc("/buckets", h.BucketsViewHandler).
		Methods(http.MethodGet)

	r.PathPrefix("/buckets/").
		HandlerFunc(h.BucketViewHandler).
		Methods(http.MethodGet)

	r.HandleFunc("/api/buckets", h.CreateBucketHandler).
		Methods(http.MethodPost)

	r.HandleFunc("/api/buckets/{bucketName}", h.DeleteBucketHandler).
		Methods(http.MethodDelete)

	r.HandleFunc("/api/buckets/{bucketName}/objects", h.CreateObjectHandler).
		Methods(http.MethodPost)

	r.HandleFunc(
		"/api/buckets/{bucketName}/objects/{objectName:.*}/url",
		h.GenerateURLHandler,
	).Methods(http.MethodGet)

	r.HandleFunc(
		"/api/buckets/{bucketName}/objects/{objectName:.*}",
		h.GetObjectHandler,
	).Methods(http.MethodGet)

	r.HandleFunc(
		"/api/buckets/{bucketName}/objects/{objectName:.*}",
		h.DeleteObjectHandle,
	).Methods(http.MethodDelete)

	return r
}
