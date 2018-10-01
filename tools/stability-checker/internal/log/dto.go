package log

// Entry represents log entry
type Entry struct {
	Level string `json:"level"`
	Log   struct {
		Message   string `json:"message"`
		TestRunID string `json:"test-run-id"`
		Time      string `json:"time"`
		Type      string `json:"type"`
	} `json:"log"`
}
