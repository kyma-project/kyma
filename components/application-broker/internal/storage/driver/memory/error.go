package memory

type notFoundError struct{}

func (notFoundError) Error() string  { return "element not found" }
func (notFoundError) NotFound() bool { return true }

type alreadyExistsError struct{}

func (alreadyExistsError) Error() string       { return "element already exists" }
func (alreadyExistsError) AlreadyExists() bool { return true }

type activeOperationInProgressError struct{}

func (activeOperationInProgressError) Error() string {
	return "there is an active operation in progres for instance"
}
func (activeOperationInProgressError) ActiveOperationInProgress() bool { return true }
