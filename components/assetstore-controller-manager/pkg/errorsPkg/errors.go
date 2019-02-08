package errorsPkg

type MultiError struct {
	message string
	errors  []error
}

func (e *MultiError) Error() string {
	return e.message
}

func (e *MultiError) Errors() []error {
	return e.errors
}

func NewMultiError(message string, errors []error) error {
	return &MultiError{
		message: message,
		errors:  errors,
	}
}

func IsMultiError(err error) bool {
	type errorWithErrors interface {
		Errors() []error
	}

	switch _ := err.(type) {
	case errorWithErrors:
		return true
	}

	return false
}
