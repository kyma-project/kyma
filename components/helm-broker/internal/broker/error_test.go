package broker_test

type notFoundError struct{}

func (notFoundError) Error() string  { return "element not found" }
func (notFoundError) NotFound() bool { return true }
