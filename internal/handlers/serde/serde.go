package serde

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/hossein1376/s3manager/pkg/errs"
	"github.com/hossein1376/s3manager/pkg/reqid"
	"github.com/hossein1376/s3manager/pkg/slogger"
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

	var e errs.Error
	if errors.As(err, &e) {
		status := e.HTTPStatusCode()
		msg := e.Message()
		slogger.Debug(
			ctx,
			"Error response",
			slogger.Err("error", e.Unwrap()),
			slog.Int("status_code", status),
			slog.String("message", msg),
		)
		WriteJson(ctx, w, status, Response{Message: msg})
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
		resp := RemoteResponse{Server: "MinIo", Message: "failed to connect"}
		WriteJson(ctx, w, http.StatusBadGateway, resp)
		return
	}

	InternalErrWrite(ctx, w, err)
}

func InternalErrWrite(ctx context.Context, w http.ResponseWriter, err error) {
	slog.ErrorContext(ctx, "Internal error", slog.Any("error", err))
	resp := Response{Message: http.StatusText(http.StatusInternalServerError)}
	id, ok := reqid.RequestID(ctx)
	if ok {
		resp.Message = id
	}
	WriteJson(ctx, w, http.StatusInternalServerError, resp)
}
