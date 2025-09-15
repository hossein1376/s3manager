package errs

type Options func(*Error)

func WithErr(err error) Options {
	return func(e *Error) {
		e.err = err
	}
}

func WithMsg(msg string) Options {
	return func(e *Error) {
		e.message = msg
	}
}

func WithHTTPStatus(statusCode int) Options {
	return func(e *Error) {
		e.httpStatusCode = statusCode
	}
}
