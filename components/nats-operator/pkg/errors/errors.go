package errors

// compile-time check
var (
	_ error       = &recoverableError{}
	_ recoverable = &recoverableError{}
)

// recoverable is an interface of an error state that the application can recover from.
type recoverable interface {
	recover()
}

// IsRecoverable returns true if the error is recoverable, otherwise returns false.
func IsRecoverable(err error) bool {
	if err == nil {
		return true
	}
	_, ok := err.(recoverable)
	return ok
}

// recoverableError represents a recoverable error.
type recoverableError struct {
	err error
}

// Recoverable returns a new recoverable error instance.
func Recoverable(err error) error {
	return &recoverableError{err: err}
}

// recover implements the recoverable interface.
func (e *recoverableError) recover() {
}

// Error implements the error interface.
func (e *recoverableError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}
