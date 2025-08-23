package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/encrypt"
)

func (h *Handler) CreateObjectHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := mux.Vars(r)["bucketName"]

	err := r.ParseMultipartForm(32 << 20) // 32 Mb
	if err != nil {
		handleHTTPError(w, fmt.Errorf("parsing multipart form: %w", err))
		return
	}
	file, _, err := r.FormFile("file")
	path := r.FormValue("path")
	if err != nil {
		handleHTTPError(w, fmt.Errorf("getting file from form: %w", err))
		return
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Printf("error closing file: %v", cErr)
		}
	}()

	opts := minio.PutObjectOptions{ContentType: "application/octet-stream"}

	if h.cfg.SSE.Type == "KMS" {
		opts.ServerSideEncryption, _ = encrypt.NewSSEKMS(h.cfg.SSE.Key, nil)
	}

	if h.cfg.SSE.Type == "SSE" {
		opts.ServerSideEncryption = encrypt.NewSSE()
	}

	if h.cfg.SSE.Type == "SSE-C" {
		opts.ServerSideEncryption, err = encrypt.NewSSEC([]byte(h.cfg.SSE.Key))
		if err != nil {
			handleHTTPError(w, fmt.Errorf("setting SSE-C key: %w", err))
			return
		}
	}

	_, err = h.s3.PutObject(r.Context(), bucketName, path, file, -1, opts)
	if err != nil {
		handleHTTPError(w, fmt.Errorf("putting object: %w", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
