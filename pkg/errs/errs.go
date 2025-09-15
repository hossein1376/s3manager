package errs

import (
	"fmt"
	"net/http"
)

type Error struct {
	err            error
	httpStatusCode int
	message        string
}

func NewErr(httpStatusCode int, opts []Options) Error {
	e := Error{httpStatusCode: httpStatusCode}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

func (e Error) Error() string {
	var text string
	if e.err != nil {
		text = e.err.Error()
	} else {
		text = http.StatusText(e.httpStatusCode)
	}
	return fmt.Sprintf("[%d] %s", e.httpStatusCode, text)
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) HTTPStatusCode() int {
	if e.httpStatusCode == 0 {
		panic("errs: http status code is zero")
	}
	return e.httpStatusCode
}

func (e Error) Message() string {
	if e.message == "" {
		return http.StatusText(e.httpStatusCode)
	}
	return e.message
}

func BadRequest(opts ...Options) Error {
	return NewErr(http.StatusBadRequest, opts)

}

func Unauthorized(opts ...Options) Error {
	return NewErr(http.StatusUnauthorized, opts)
}

func Forbidden(opts ...Options) Error {
	return NewErr(http.StatusForbidden, opts)
}

func NotFound(opts ...Options) Error {
	return NewErr(http.StatusNotFound, opts)
}

func Conflict(opts ...Options) Error {
	return NewErr(http.StatusConflict, opts)
}

func TooMany(opts ...Options) Error {
	return NewErr(http.StatusTooManyRequests, opts)
}

func Internal(opts ...Options) Error {
	return NewErr(http.StatusInternalServerError, opts)
}

func Timeout(opts ...Options) Error {
	return NewErr(http.StatusGatewayTimeout, opts)
}
