package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/hossein1376/s3manager/internal/handlers/serde"
)

func (h *Handler) GenerateURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	bucketName := strings.TrimSpace(vars["bucketName"])
	objectName := strings.TrimSpace(vars["objectName"])
	expiry := r.URL.Query().Get("expiry")
	if bucketName == "" || objectName == "" {
		resp := serde.Response{
			Message: "bucket name and object name must be specified",
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	parsedExpiry, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil {
		resp := serde.Response{
			Message: fmt.Errorf("converting expiry: %w", err).Error(),
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	if parsedExpiry > 7*24*60*60 || parsedExpiry < 1 {
		resp := serde.Response{
			Message: fmt.Sprintf("invalid expiry value: %d", parsedExpiry),
		}
		serde.WriteJson(ctx, w, http.StatusBadRequest, resp)
		return
	}

	expirySecond := time.Duration(parsedExpiry * 1e9)
	u, err := h.s3.PresignedGetObject(
		ctx, bucketName, objectName, expirySecond, make(url.Values),
	)
	if err != nil {
		serde.ExtractAndWrite(ctx, w, fmt.Errorf("generating url: %w", err))
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(map[string]string{"url": u.String()})
	if err != nil {
		serde.InternalErrWrite(ctx, w, fmt.Errorf("encoding JSON: %w", err))
		return
	}
}
