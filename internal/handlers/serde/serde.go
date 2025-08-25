package serde

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/minio/minio-go/v7"
)

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type RemoteResponse struct {
	Server     string `json:"server"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
}

func ExtractAndWrite(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		WriteJson(ctx, w, http.StatusNoContent, nil)
		return
	}

	var e minio.ErrorResponse
	if errors.As(err, &e) {
		resp := RemoteResponse{
			Server: e.Server, StatusCode: e.StatusCode, Message: e.Message,
		}
		WriteJson(ctx, w, http.StatusBadGateway, resp)
		return
	}

	var timeoutErr net.Error
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		WriteJson(
			ctx, w, http.StatusGatewayTimeout, RemoteResponse{Server: "MinIo"},
		)
		return
	}
	var opError *net.OpError
	if errors.As(err, &opError) && errors.Is(opError.Err, syscall.ECONNREFUSED) {
		resp := RemoteResponse{Server: "MinIo", Message: "server is down"}
		WriteJson(ctx, w, http.StatusBadGateway, resp)
		return
	}

	InternalErrWrite(ctx, w, err)
}

func InternalErrWrite(ctx context.Context, w http.ResponseWriter, err error) {
	slog.ErrorContext(ctx, "Internal error", slog.Any("error", err))
	WriteJson(
		ctx,
		w,
		http.StatusInternalServerError,
		http.StatusText(http.StatusInternalServerError),
	)
}
