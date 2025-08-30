package handlers

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"

	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/web"
)

const (
	BucketName = "bucketName"
	ObjectName = "objectName"
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

func newRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Handle(
		"GET /", http.RedirectHandler("/buckets", http.StatusPermanentRedirect),
	)
	r.Handle(
		"GET /static/*",
		http.StripPrefix("/static/", http.FileServer(http.FS(h.statics))),
	)
	r.Get("/buckets", h.ListBucketsHandler)
	r.Get("/buckets/*", h.ViewBucketHandler)
	r.Post("/api/buckets", h.CreateBucketHandler)
	r.Delete("/api/buckets/{bucketName}", h.DeleteBucketHandler)
	r.Post("/api/buckets/{bucketName}/objects", h.CreateObjectHandler)
	r.Get(
		"/api/buckets/{bucketName}/objects/{objectName}/url",
		h.GenerateURLHandler,
	)
	r.Get(
		"/api/buckets/{bucketName}/objects/{objectName}",
		h.GetObjectHandler,
	)
	r.Delete(
		"/api/buckets/{bucketName}/objects/{objectName}",
		h.DeleteObjectHandle,
	)

	return r
}
