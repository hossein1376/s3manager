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

func ExtractAndWrite(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		WriteJson(ctx, w, http.StatusNoContent, nil)
		return
	}

	var e errs.Error
	if errors.As(err, &e) {
		status := e.HTTPStatusCode()
		msg := e.Message()
		if msg == "" {
			msg = e.Unwrap().Error()
		}
		slogger.Debug(
			ctx,
			"Error response",
			slogger.Err("error", err),
			slog.Int("status_code", status),
			slog.String("message", e.Message()),
		)
		WriteJson(ctx, w, status, Response{Message: msg})
		return
	}

	var timeoutErr net.Error
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		slogger.Warn(ctx, "server timeout", slogger.Err("error", err))
		resp := Response{Message: "server timeout"}
		WriteJson(ctx, w, http.StatusGatewayTimeout, resp)
		return
	}
	var opError *net.OpError
	if errors.As(err, &opError) && errors.Is(opError.Err, syscall.ECONNREFUSED) {
		slogger.Warn(ctx, "server connection refused", slogger.Err("error", err))
		resp := Response{Message: "failed to connect to server"}
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
		resp.Data = id
	}
	WriteJson(ctx, w, http.StatusInternalServerError, resp)
}
