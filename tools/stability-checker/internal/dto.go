package internal

// TestStatus holds necessary information about given test execution
type TestStatus struct {
	ID   string `json:"id"`
	Pass bool   `json:"pass"`
}
