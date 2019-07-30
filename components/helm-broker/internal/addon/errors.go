package addon

import "github.com/pkg/errors"

// Kind represents the kind of the error
type Kind int

const (
	// Unknown represents unknown error about loading addon entry
	Unknown Kind = iota
	// LoadingErrorKind represents error occurred when addon
	// was successfully fetched from external repository but processing/loading phase failed
	LoadingErrorKind
	// FetchingErrorKind represents error occurred when addon cannot be fetched from external repository
	FetchingErrorKind
)

// String returns name of the given kind
func (r Kind) String() string {
	switch r {
	case LoadingErrorKind:
		return "LoadingError"
	case FetchingErrorKind:
		return "FetchingError"
	case Unknown:
		return "UnknownError"
	default:
		return ""
	}
}

var _ error = &Error{}

// Error holds information about the addon error
type Error struct {
	kind Kind
	err  error
}

// Error returns error message
func (e *Error) Error() string {
	return e.err.Error()
}

// Kind returns error kind
func (e *Error) Kind() Kind {
	return e.kind
}

// NewLoadingError returns new loading error
func NewLoadingError(err error) error {
	return &Error{
		kind: LoadingErrorKind,
		err:  err,
	}
}

// NewFetchingError returns new fetching error
func NewFetchingError(err error) error {
	return &Error{
		kind: FetchingErrorKind,
		err:  err,
	}
}

// IsLoadingError returns true only if given error represents the loading error kind
func IsLoadingError(err error) bool {
	return isKindError(err, LoadingErrorKind)
}

// IsFetchingError returns true only if given error represents the fetching error kind
func IsFetchingError(err error) bool {
	return isKindError(err, FetchingErrorKind)
}

func isKindError(err error, kind Kind) bool {
	err = errors.Cause(err)
	if err == nil {
		return false
	}

	return reasonForError(err) == kind
}

func reasonForError(err error) Kind {
	switch t := err.(type) {
	case *Error:
		return t.kind
	default:
		return Unknown
	}
}
