package handlers

import (
	"net/http"

	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/internal/services"
)

const (
	BucketName = "bucketName"
	ObjectName = "objectName"
)

type Handler struct {
	cfg     *config.Config
	service *services.Services
}

func NewServer(
	cfg *config.Config, srvc *services.Services) *http.Server {
	h := &Handler{cfg: cfg, service: srvc}

	return &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      newRouter(h),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
}

func newRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(
		"GET /", http.RedirectHandler("/buckets", http.StatusPermanentRedirect),
	)
	mux.HandleFunc("GET /buckets", h.ListBucketsHandler)
	mux.HandleFunc("GET /buckets/{bucket}", h.ListObjectsHandler)
	mux.HandleFunc("POST /buckets", h.CreateBucketHandler)
	mux.HandleFunc("DELETE /buckets/{bucket}", h.DeleteBucketHandler)
	mux.HandleFunc("PUT /buckets/{bucket}/objects", h.PutObjectHandler)
	mux.HandleFunc(
		"GET /buckets/{bucket}/objects/{object}", h.GetObjectHandler,
	)
	mux.HandleFunc(
		"DELETE /buckets/{bucket}/objects/{object}", h.DeleteObjectHandle,
	)

	return mux
}
