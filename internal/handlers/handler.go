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
		"GET /", http.FileServer(http.Dir("./ui")),
	)
	mux.Handle("GET /api/buckets", withDefaults(h.ListBucketsHandler))
	mux.Handle("GET /api/buckets/{bucket}", withDefaults(h.ListObjectsHandler))
	mux.Handle("POST /api/buckets", withDefaults(h.CreateBucketHandler))
	mux.Handle(
		"DELETE /api/buckets/{bucket}", withDefaults(h.DeleteBucketHandler),
	)
	mux.Handle(
		"PUT /api/buckets/{bucket}/objects", withDefaults(h.PutObjectHandler),
	)
	mux.Handle(
		"GET /api/buckets/{bucket}/objects/{object}",
		withDefaults(h.GetObjectHandler),
	)
	mux.Handle(
		"DELETE /api/buckets/{bucket}/objects/{object}",
		withDefaults(h.DeleteObjectHandle),
	)

	return mux
}

func withDefaults(handler http.HandlerFunc) http.Handler {
	return withMiddlewares(
		handler,
		requestIDMiddleware,
		loggerMiddleware,
		recoverMiddleware,
		corsMiddleware,
	)
}
