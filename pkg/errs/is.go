package errs

import (
	"errors"
	"net/http"
)

func Is(err error) (Error, bool) {
	var e Error
	return e, errors.As(err, &e)
}

func IsForbidden(err error) bool {
	var e Error
	return errors.As(err, &e) && e.httpStatusCode == http.StatusForbidden
}

func IsNotFound(err error) bool {
	var e Error
	return errors.As(err, &e) && e.httpStatusCode == http.StatusNotFound
}

func IsConflict(err error) bool {
	var e Error
	return errors.As(err, &e) && e.httpStatusCode == http.StatusConflict
}

func IsTooManyReqs(err error) bool {
	var e Error
	return errors.As(err, &e) && e.httpStatusCode == http.StatusTooManyRequests
}
