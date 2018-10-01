package internal

// ExecutionStatus holds necessary information about given test execution
type ExecutionStatus struct {
	ID   string `json:"id"`
	Pass bool   `json:"pass"`
}
