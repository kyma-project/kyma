package httpclient

type Error struct {
	StatusCode int
	Message    string
	Cause      error
}

type ErrorOpt func(*Error)

func NewError(err error, opts ...ErrorOpt) *Error {
	e := &Error{Cause: err}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Cause
}

func WithStatusCode(statusCode int) ErrorOpt {
	return func(e *Error) {
		e.StatusCode = statusCode
	}
}

func WithMessage(message string) ErrorOpt {
	return func(e *Error) {
		e.Message = message
	}
}
