package reqid

import (
	"context"
	"crypto/rand"
	"log/slog"

	"github.com/oklog/ulid/v2"
)

type ReqID string

const RequestIDKey ReqID = "request_id"

func NewRequestID() ReqID {
	id, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		slog.Error("Generate request id", slog.Any("error", err))
	}
	return ReqID(id.String())
}

func RequestID(c context.Context) (string, bool) {
	id, ok := c.Value(RequestIDKey).(ReqID)
	return string(id), ok
}
