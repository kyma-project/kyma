package broker

// IsNotFoundError check if error is NotFound one.
func IsNotFoundError(err error) bool {
	nfe, ok := err.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

// IsForbiddenError checks if error represent Forbidden one.
func IsForbiddenError(err error) bool {
	type forbidden interface {
		Forbidden() bool
	}

	if t, ok := err.(forbidden); ok {
		return t.Forbidden()
	}
	return false
}

// ForbiddenError represents situation when operation is forbidden
type ForbiddenError struct {
}

func (fe *ForbiddenError) Error() string {
	return "Forbidden Error"
}

// Forbidden is a marker method, used in IsForbiddenError method
func (fe *ForbiddenError) Forbidden() bool {
	return true
}
