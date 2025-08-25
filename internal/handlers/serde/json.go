package serde

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// WriteJson will write back data in json format with the provided status code
// and headers.
func WriteJson(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	data any,
) {
	if data == nil {
		w.WriteHeader(status)
		return
	}

	js, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(
			ctx, "Marshal data", slog.Any("data", data), slog.Any("error", err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		slog.ErrorContext(
			ctx, "Write", slog.Any("data", data), slog.Any("error", err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

// ReadJson will decode incoming json requests. It will return a human-readable
// error in case of failure.
func ReadJson(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err == nil {
		if err = dec.Decode(&struct{}{}); err != io.EOF {
			return errors.New("body must only contain a single JSON value")
		}
		return nil
	}

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	switch {
	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")
	case errors.As(err, &syntaxError):
		return fmt.Errorf(
			"body contains badly-formed JSON (at character %d)",
			syntaxError.Offset,
		)
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf(
				"body contains incorrect JSON type for field %q",
				unmarshalTypeError.Field,
			)
		}
		return fmt.Errorf(
			"body contains incorrect JSON type (at character %d)",
			unmarshalTypeError.Offset,
		)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("body contains unknown key %s", fieldName)
	default:
		return err
	}
}
