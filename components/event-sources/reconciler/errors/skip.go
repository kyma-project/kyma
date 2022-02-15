package errors

// skippable is a type of error that can skip reconciliation retries.
type skippable interface {
	Skip()
}

// IsSkippable returns whether the given error is skippable.
func IsSkippable(e error) bool {
	_, ok := e.(skippable)
	return ok
}

// NewSkippable wraps an error into a skippable type so it gets ignored by the
// reconciler.
func NewSkippable(e error) error {
	return &skipRetry{err: e}
}

type skipRetry struct{ err error }

// type skipRetry implements skippable.
var _ skippable = &skipRetry{}

// Skip implements skippable.
func (*skipRetry) Skip() {}

// Error implements the error interface.
func (e *skipRetry) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}
