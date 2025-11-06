package handlers

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/internal/services"
	"github.com/hossein1376/s3manager/ui"
)

type Handler struct {
	cfg     config.Config
	service *services.Services
}

func NewServer(
	cfg config.Config, svc *services.Services,
) (*http.Server, error) {
	h := &Handler{cfg: cfg, service: svc}

	uiFS, err := ui.FileSystem()
	if err != nil {
		return nil, fmt.Errorf("loading ui filesystem: %w", err)
	}

	return &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      newRouter(h, http.FS(uiFS), cfg.Server.DisableUI),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}, nil
}

func newRouter(h *Handler, ui http.FileSystem, disableUI bool) *grape.Router {
	r := grape.NewRouter()
	r.UseAll(
		grape.RequestIDMiddleware,
		grape.LoggerMiddleware,
		grape.RecoverMiddleware,
		grape.CORSMiddleware,
	)

	if !disableUI {
		r.Get("/", toHandlerFunc(http.FileServer(ui)))
	}
	r.Get("/api/buckets", h.ListBucketsHandler)
	r.Post("/api/buckets", h.CreateBucketHandler)
	r.Get("/api/buckets/{bucket}", h.ListObjectsHandler)
	r.Delete("/api/buckets/{bucket}", h.DeleteBucketHandler)
	r.Put("/api/buckets/{bucket}/objects", h.PutObjectHandler)
	r.Get("/api/buckets/{bucket}/objects/{object}", h.GetObjectHandler)
	r.Delete("/api/buckets/{bucket}/objects/{object}", h.DeleteObjectHandle)

	return r
}

func toHandlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
