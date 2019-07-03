package bundle

const (
	FetchingErrorKind   = "FetchingError"
	ValidationErrorKind = "ValidationError"
)

type BundleError interface {
	Error() string
	Kind() string
}

type FetchingError struct {
	kind string
	err  error
}

func NewFetchingError(msg error) *FetchingError {
	return &FetchingError{
		kind: FetchingErrorKind,
		err:  msg,
	}
}

func (e *FetchingError) Error() string { return e.err.Error() }

func (e *FetchingError) Kind() string { return e.kind }

type ValidationError struct {
	kind string
	msg  error
}

func NewValidationError(err error) *ValidationError {
	return &ValidationError{
		kind: ValidationErrorKind,
		msg:  err,
	}
}

func (e *ValidationError) Error() string { return e.msg.Error() }

func (e *ValidationError) Kind() string { return e.kind }

func IsFetchingError(err error) bool {
	return ReasonForError(err) == FetchingErrorKind
}

func IsValidationError(err error) bool {
	return ReasonForError(err) == ValidationErrorKind
}

func ReasonForError(err error) string {
	switch t := err.(type) {
	default:
		return ""
	case *ValidationError:
		return t.kind
	case *FetchingError:
		return t.kind
	}
}
